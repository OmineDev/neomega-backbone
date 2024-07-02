package sync_wrapper

import (
	"sync"
	"sync/atomic"
)

// 基于go1.22.2的sync.map包修改
/*
原有方法不变, 添加了注释，新增了下述原子方法：
SwapMultiple, 保存多个键并返回旧值
LoadMultiple, 读取多个目标键
DeleteMultiple, 删除多个键并返回旧值
*/

type EnhancedMap struct {
	mu     sync.Mutex               //仅当对dirty操作时进行上锁
	read   atomic.Pointer[readOnly] //只读表，指不能更改键数量，但是可以更改键的值
	dirty  map[any]*entry           //读写表，操作时要加锁
	misses int                      //read不命中的次数到达len(dirty)时，dirty将替换成为新的read
}

// 此表一旦生成不可以再更改，要更改必须建立新的readOnly并使用原子方法保存来替换掉
type readOnly struct {
	m       map[any]*entry
	amended bool //dirty中是否有更多的键
}

/*
硬删除标记，区别于软删除nil；
当键在read中被删除时，首先会被置为nil软删除，此时键仍然存在；
当dirty被提升为read后，dirty将整体被置为nil，然后再由read回写到dirty，
此时若回写的值为nil则不回写，并将nil标记换为硬删除标记expunged，表示dirty中不记录这个键；
expunged的存在使得键在read中删除和恢复更快速，对硬删除键要赋值必须先转为软删除状态（即在dirty中创建相应键）
expunged标记的键的完全清除将发生在dirty提升为read时自然抛弃
*/
var expunged = new(any)

type entry struct {
	p atomic.Pointer[any]
}

func newEntry(i any) *entry {
	e := &entry{}
	e.p.Store(&i)
	return e
}

// 因为ReadOnly随时可能被更改，用原子性保证获取到正确以及最新的ReadOnly
func (m *EnhancedMap) loadReadOnly() readOnly {
	if p := m.read.Load(); p != nil {
		return *p
	}
	return readOnly{}
}

func (m *EnhancedMap) Load(key any) (value any, ok bool) {
	read := m.loadReadOnly()
	e, ok := read.m[key]
	if !ok && read.amended { //未命中，前往dirty
		m.mu.Lock()
		// Doublr Checking
		read = m.loadReadOnly()
		e, ok = read.m[key]
		if !ok && read.amended { //未命中，前往dirty
			e, ok = m.dirty[key]
			m.missLocked() // 不管条目是否存在，记录一个未命中
		}
		m.mu.Unlock()
	}
	if !ok {
		return nil, false
	}
	return e.load()
}

// 软删除或者硬删除状态下不能加载值
func (e *entry) load() (value any, ok bool) {
	p := e.p.Load()
	if p == nil || p == expunged {
		return nil, false
	}
	return *p, true
}

// Store和Swap一样
func (m *EnhancedMap) Store(key, value any) {
	_, _ = m.Swap(key, value)
}

func (e *entry) tryCompareAndSwap(old, new any) bool {
	p := e.p.Load()
	if p == nil || p == expunged || *p != old {
		return false
	}
	/*
		通过建立副本的方式告诉编译器new不会在tryCompareAndSwap中被传递出去
		让编译器更倾向于将new分配在栈上而不是堆上
		作为副本承担逃逸风险的的nc也因为其更小的生命周期和作用域更容易被分配在栈上
		总体上提升了被分配到栈的权重，对原代码逻辑无影响
	*/
	nc := new
	for {
		if e.p.CompareAndSwap(p, &nc) {
			return true
		}
		p = e.p.Load()
		if p == nil || p == expunged || *p != old {
			return false
		}
		// 运行到这在竞态中符合上述两种情况的代码在被判断之前受到了外部更改导致判断失效，要循环执行直至成功
	}
}

// 如果entry当前是硬删状态，则从硬删除状态转为软删除状态
func (e *entry) unexpungeLocked() (wasExpunged bool) {
	return e.p.CompareAndSwap(expunged, nil)
}

func (e *entry) swapLocked(i *any) *any {
	return e.p.Swap(i)
}

func (m *EnhancedMap) LoadOrStore(key, value any) (actual any, loaded bool) {
	read := m.loadReadOnly()
	if e, ok := read.m[key]; ok { //键存在则尝试进行LOS
		actual, loaded, ok := e.tryLoadOrStore(value)
		if ok {
			return actual, loaded
		}
	}

	//运行到这说明read没有这个键，无论是读还是存都要进dirty，因此不用检查amended

	m.mu.Lock()
	read = m.loadReadOnly()
	//double checking
	if e, ok := read.m[key]; ok { //这个键在两个表中都存在
		if e.unexpungeLocked() { //如果e.p值是强删除状态,那么需要改为软删除状态才能存值
			m.dirty[key] = e //软删除状态，即双map键存在，键e.p值为nil
		}
		actual, loaded, _ = e.tryLoadOrStore(value)
	} else if e, ok := m.dirty[key]; ok { //仅在dirty表
		actual, loaded, _ = e.tryLoadOrStore(value)
		m.missLocked()
	} else { //俩表都不存在
		if !read.amended {
			m.dirtyLocked()
			m.read.Store(&readOnly{m: read.m, amended: true})
		}
		m.dirty[key] = newEntry(value)
		actual, loaded = value, false
	}
	m.mu.Unlock()

	return actual, loaded
}

// 返回值：
// ok：操作是否成功（Load 成功、Store 成功）
// loaded：表示是否是 Load 出来的
// actual：Load 到的值
func (e *entry) tryLoadOrStore(i any) (actual any, loaded, ok bool) {
	p := e.p.Load()
	if p == expunged { //已经被硬删除，操作失败
		return nil, false, false
	}
	if p != nil { //Load成功
		return *p, true, true
	}

	// 运行到这里说明是nil软删除状态，需要Store

	/*
		通过建立副本的方式告诉编译器i不会在tryLoadOrStore中被传递出去
		让编译器将更倾向于将i分配在栈上而不是堆上
		作为副本的ic也因为其更小的生命周期和作用域更容易被分配在栈上
		总体上有优化，对原代码逻辑无影响
	*/
	ic := i
	for {
		if e.p.CompareAndSwap(nil, &ic) {
			return i, false, true //是软删除状态，与nil交换成功，即成功Store
		}
		p = e.p.Load()
		if p == expunged { //是硬删除状态，Stroe失败
			return nil, false, false
		}
		if p != nil { //是正常状态，则Stroe失败，改为Load成功
			return *p, true, true
		}
	}
}

// 删除键的值，返回以前的值（如果有）
// loaded 表示 key 是否存在
func (m *EnhancedMap) LoadAndDelete(key any) (value any, loaded bool) {
	read := m.loadReadOnly()
	e, ok := read.m[key]
	if !ok && read.amended { //read找不到
		m.mu.Lock()
		// double checking
		read = m.loadReadOnly()
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = m.dirty[key]
			delete(m.dirty, key) //仅在dirty中存在的键直接删除，在read也存在才会经过软删除和硬删除再删除以节省开销
			m.missLocked()
		}
		m.mu.Unlock()
	}
	if ok {
		return e.delete()
	}
	return nil, false
}

func (m *EnhancedMap) Delete(key any) {
	m.LoadAndDelete(key)
}

func (e *entry) delete() (value any, ok bool) {
	for {
		p := e.p.Load()
		if p == nil || p == expunged { //删无可删
			return nil, false
		}
		if e.p.CompareAndSwap(p, nil) {
			return *p, true
		}
	}
}

func (e *entry) trySwap(i *any) (*any, bool) {
	for {
		// 原子操作获取键状态
		p := e.p.Load()
		if p == expunged {
			return nil, false //键已经被强删除，交换失败
		}
		if e.p.CompareAndSwap(p, i) {
			return p, true
		}
	}
}

func (m *EnhancedMap) Swap(key, value any) (previous any, loaded bool) {
	read := m.loadReadOnly()
	if e, ok := read.m[key]; ok {
		if v, ok := e.trySwap(&value); ok {
			if v == nil {
				return nil, false
			}
			return *v, true
		}
	}
	// 运行到这就是read中没有键或者键状态为强删除expunged
	m.mu.Lock()
	// Double Checking
	read = m.loadReadOnly()
	if e, ok := read.m[key]; ok {
		if e.unexpungeLocked() {
			m.dirty[key] = e
		}

		// 运行到这里e.p值要么是有效值，要么是nil，都是可以直接赋值的状态

		if v := e.swapLocked(&value); v != nil { //e.p原子操作Swap
			// 如果是有效值
			loaded = true
			previous = *v
		}
	} else if e, ok := m.dirty[key]; ok { //这个键仅在dirty表中存在
		if v := e.swapLocked(&value); v != nil {
			loaded = true
			previous = *v
		}
	} else {
		// 运行到这里代表这个键在两个表都不存在
		if !read.amended {
			m.dirtyLocked()                                   //如果dirty为nil则初始化
			m.read.Store(&readOnly{m: read.m, amended: true}) //更换readOnly以更改amended
		}
		m.dirty[key] = newEntry(value)
	}
	m.mu.Unlock()
	return previous, loaded
}

func (m *EnhancedMap) CompareAndSwap(key, old, new any) bool {
	read := m.loadReadOnly()
	if e, ok := read.m[key]; ok {
		return e.tryCompareAndSwap(old, new) //read中存在则直接CAS
	} else if !read.amended {
		return false // dirty中也没有这个键
	}
	// 执行到这是去dirty中确认是否有目标键并进行CAS
	m.mu.Lock()
	defer m.mu.Unlock()
	//double checking
	read = m.loadReadOnly()
	swapped := false
	if e, ok := read.m[key]; ok {
		swapped = e.tryCompareAndSwap(old, new)
	} else if e, ok := m.dirty[key]; ok {
		swapped = e.tryCompareAndSwap(old, new)
		m.missLocked()
	}
	return swapped
}

func (m *EnhancedMap) CompareAndDelete(key, old any) (deleted bool) {
	read := m.loadReadOnly()
	e, ok := read.m[key]
	if !ok && read.amended {
		m.mu.Lock()
		read = m.loadReadOnly()
		e, ok = read.m[key]
		if !ok && read.amended { //read未命中及其double checking
			e, ok = m.dirty[key]
			m.missLocked()
		}
		m.mu.Unlock()
	}
	for ok {
		p := e.p.Load()
		//条件判断
		if p == nil || p == expunged || *p != old {
			return false
		}
		//设为软删除
		if e.p.CompareAndSwap(p, nil) {
			return true
		}
		//防止竞态下对判断条件的错过，要加循环直至成功执行
	}
	return false
}

func (m *EnhancedMap) Range(f func(key, value any) bool) {
	read := m.loadReadOnly()
	if read.amended { //遍历先将dirty提升为read再遍历
		m.mu.Lock()
		read = m.loadReadOnly()
		if read.amended {
			read = readOnly{m: m.dirty}
			copyRead := read
			m.read.Store(&copyRead)
			m.dirty = nil
			m.misses = 0
		}
		m.mu.Unlock()
	}

	for k, e := range read.m {
		v, ok := e.load() //硬删除或者软删除则跳过
		if !ok {
			continue
		}
		if !f(k, v) { // f 可以返回一个 bool 值，如果返回 false，那么就停止遍历
			break
		}
	}
}

func (m *EnhancedMap) missLocked() {
	m.misses++
	if m.misses < len(m.dirty) {
		return
	}
	// 未命中的次数达到 len(m.dirty)，将 dirty map 提升为 read map
	m.read.Store(&readOnly{m: m.dirty})
	// 重置 dirty map
	m.dirty = nil
	// 重置 misses
	m.misses = 0
}

// 如果 m.dirty 为 nil，则创建一个新的 dirty map，否则不做任何操作
func (m *EnhancedMap) dirtyLocked() {
	if m.dirty != nil {
		return
	}

	read := m.loadReadOnly()
	m.dirty = make(map[any]*entry, len(read.m))
	for k, e := range read.m {
		// read map 中 nil 的 key 会被转换为 expunged 状态
		// 对于 read map 中的 key，如果不是 expunged，则将其复制到 dirty map 中
		if !e.tryExpungeLocked() {
			m.dirty[k] = e
		}
	}
}

func (e *entry) tryExpungeLocked() (isExpunged bool) {
	p := e.p.Load()
	for p == nil {
		if e.p.CompareAndSwap(nil, expunged) {
			return true
		}
		p = e.p.Load()
	}
	return p == expunged
}

// Multi类专用，因为read是键的快照，不是键值的快照，Multi操作必须在dirty中进行
// 将dirty中非传入的键提升为read,必须先lock dirty才能调用
func (m *EnhancedMap) updateReadWithoutKeysLocked(aim_keys map[any]any) {
	m.dirtyLocked()
	//根据传入键列表分为两份
	withoutAimKeysMap := make(map[any]*entry, len(m.dirty))
	haveAimKeysMap := make(map[any]*entry, len(aim_keys))
	for k, e := range m.dirty {
		if _, ok := aim_keys[k]; ok {
			haveAimKeysMap[k] = e
		} else {
			withoutAimKeysMap[k] = e
		}
	}

	// 不包含部分提升为read,因为只是部分提升，amended为true
	m.read.Store(&readOnly{m: withoutAimKeysMap, amended: true})
	// 重置
	m.misses = 0

	read := m.loadReadOnly()
	m.dirty = make(map[any]*entry, len(read.m)+len(haveAimKeysMap))
	for k, e := range read.m {
		// read map 中 nil 的 key 会被转换为 expunged 状态
		// 对于 read map 中的 key，如果不是 expunged，则将其复制到 dirty map 中
		if !e.tryExpungeLocked() {
			m.dirty[k] = e
		}
	}
	for k, e := range haveAimKeysMap {
		m.dirty[k] = e
	}
}

// 新方法：保存多个键并返回
// 上锁，将要保存的键从read中退回，然后在dirty中统一写入，再提升到read（反正锁都锁了）
func (m *EnhancedMap) SwapMultiple(pairs map[any]any) map[any]any {
	old := make(map[any]any, len(pairs))
	if len(pairs) == 0 {
		return old
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateReadWithoutKeysLocked(pairs)
	//统一换
	for key, value := range pairs {
		var previous any
		if e, ok := m.dirty[key]; ok {
			if v := e.swapLocked(&value); v != nil {
				previous = *v
			}
		} else {
			m.dirty[key] = newEntry(value)
		}
		old[key] = previous
	}
	// 提升dirty到read
	m.read.Store(&readOnly{m: m.dirty})
	m.misses = 0
	read := m.loadReadOnly()
	m.dirty = make(map[any]*entry, len(read.m))
	for k, e := range read.m {
		if !e.tryExpungeLocked() {
			m.dirty[k] = e
		}
	}
	return old
}

// 新方法：读取多个键
// 上锁，将要读取的键从read中退回，然后在dirty中统一读取，再提升到read（反正锁都锁了）
func (m *EnhancedMap) LoadMultiple(keys []any) map[any]any {
	result := make(map[any]any, len(keys))
	if len(keys) == 0 {
		return result
	}
	aimKeysSet := make(map[any]any, len(keys))
	for _, key := range keys {
		aimKeysSet[key] = struct{}{}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateReadWithoutKeysLocked(aimKeysSet)
	// 统一读
	for _, key := range keys {
		e, ok := m.dirty[key]
		if ok {
			if val, load_ok := e.load(); load_ok {
				result[key] = val
				continue
			}
		}
		result[key] = nil
	}
	// 提升dirty到read
	m.read.Store(&readOnly{m: m.dirty})
	m.misses = 0
	read := m.loadReadOnly()
	m.dirty = make(map[any]*entry, len(read.m))
	for k, e := range read.m {
		if !e.tryExpungeLocked() {
			m.dirty[k] = e
		}
	}
	return result
}

// 新方法：删除所有指定前缀的键
// 上锁，将要删除的键从read中退回，然后在dirty中统一删除，再提升到read（反正锁都锁了）
func (m *EnhancedMap) DeleteMultiple(keys []any) map[any]any {
	result := make(map[any]any, len(keys))
	if len(keys) == 0 {
		return result
	}
	aimKeysSet := make(map[any]any, len(keys))
	for _, key := range keys {
		aimKeysSet[key] = struct{}{}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateReadWithoutKeysLocked(aimKeysSet)
	// 统一删
	for _, key := range keys {
		e, ok := m.dirty[key]
		delete(m.dirty, key) //仅在dirty中存在的键直接删除，在read也存在才会经过软删除和硬删除再删除以节省开销
		if ok {
			val, _ := e.delete()
			result[key] = val
		}
	}
	// 提升dirty到read
	m.read.Store(&readOnly{m: m.dirty})
	m.misses = 0
	read := m.loadReadOnly()
	m.dirty = make(map[any]*entry, len(read.m))
	for k, e := range read.m {
		if !e.tryExpungeLocked() {
			m.dirty[k] = e
		}
	}
	return result
}

// 基于上述修改的新的泛型syncmap
type SyncKVEnhancedMap[K, T any] struct {
	EnhancedMap
}

func NewSyncKVEnhancedMap[K, T any]() *SyncKVEnhancedMap[K, T] {
	return &SyncKVEnhancedMap[K, T]{EnhancedMap{}}
}

func (m *SyncKVEnhancedMap[K, T]) Set(key K, value T) (previous T, loaded bool) {
	pre, loaded := m.EnhancedMap.Swap(key, value)
	if loaded {
		return pre.(T), loaded
	}
	return previous, false
}

func (m *SyncKVEnhancedMap[K, T]) Get(key K) (val T, ok bool) {
	v, ok := m.EnhancedMap.Load(key)
	if ok {
		return v.(T), true
	} else {
		return val, false
	}
}

func (m *SyncKVEnhancedMap[K, T]) Delete(key K) (val T, loaded bool) {
	value, loaded := m.EnhancedMap.LoadAndDelete(key)
	if loaded {
		return value.(T), loaded
	}
	return val, loaded
}

func (m *SyncKVEnhancedMap[K, T]) GetOrSet(key K, value T) (val T, ok bool) {
	v, ok := m.EnhancedMap.LoadOrStore(key, value)
	if ok {
		return v.(T), ok
	}
	return value, ok
}

func (m *SyncKVEnhancedMap[K, T]) GetAndDelete(key K) (val T, ok bool) {
	v, ok := m.EnhancedMap.LoadAndDelete(key)
	if ok {
		return v.(T), true
	} else {
		return val, false
	}
}

func (m *SyncKVEnhancedMap[K, T]) Iter(fn func(k K, v T) (continueInter bool)) {
	m.EnhancedMap.Range(func(key, value any) bool {
		return fn(key.(K), value.(T))
	})
}

func (m *SyncKVEnhancedMap[K, T]) CompareAndSwap(key K, old T, new T) (ok bool) {
	return m.EnhancedMap.CompareAndSwap(key, old, new)
}
