package nkdb

type Storage interface {
	Load() ([][]string, error)
	Save([][]string) error
}
