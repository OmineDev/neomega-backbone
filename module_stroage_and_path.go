package neomega_backbone

type KVDBLike interface {
	Get(key string) (value string)
	Delete(key string)
	Set(key string, value string)
	Iter(func(key, value string) bool)
}

type StorageAndPathAccess interface {
	// ${log}/topic
	GetLoggerPath(topic string) string
	// ${data}/topic
	GetFileData(topic string) ([]byte, error)
	GetJsonData(topic string, data interface{}) error
	WriteFileData(topic string, data []byte) error
	WriteJsonData(topic string, data interface{}) error
	WriteJsonDataWithTMP(topic string, tmpSuffix string, data interface{}) error
	// ${data}/topic
	GetFilePath(elem ...string) string
	// ${cache}/topic
	GetOmegaCachePath(elem ...string) string
	// ${archive}/topic
	GetArchivePath(elem ...string) string
	// ${config}/topic
	GetConfigPath(elem ...string) string
	// ${temp}/random_name
	NewTempDir() string
	// ${lang_specific}/topic, e.g. ${lang_specific}/lua, ${lang_specific}/side-python
	GetLangSpecificPath(elem ...string) string
	// on system like android, we can not use "seek" or some specific file operation under download or dirs in public dir,
	// which makes it impossible to use a normal database
	// FileLogKVDBLike is a KVDBLike, which aims to work in a file-system where "seek" is not supported
	GetNoSeekKVDBLike(elem ...string) (KVDBLike, error)
}

type StorageAndPathProvider interface {
	StorageAndPathAccess
	//
	CanPreInit
}
