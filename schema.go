package flyway

type Schema struct {
	InstalledRank int
	Version       string
	Description   string
	Script        string
	Sql           string
}
