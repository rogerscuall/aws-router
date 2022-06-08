package ports

type DbPort interface {
	CloseDbConnection()
	GetVal(key string) ([]byte, error)
	SetVal(key string, val []byte) error
	Sync()
}
