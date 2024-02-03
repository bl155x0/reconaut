package storage

type StoreHandler struct {
	ResultName  string        `yaml:"ResultName"`
	StorageName string        `yaml:"Storage"`
	Data        []StorageData `yaml:"Data"`
}

type StorageData struct {
	Column string `yaml:"Column"`
	Value  string `yaml:"Value"`
}

type StorageDefinition struct {
	Name            string  `yaml:"Name"`
	CreateStatement string  `yaml:"CreateStatement"`
	Queries         []Query `yaml:"Query"`
}

type Query struct {
	Name      string `yaml:"Name"`
	Statement string `yaml:"Statement"`
}
