package pkgs

type Index struct {
	Packages []struct {
		Name       string `json:"name"`
		Maintainer string `json:"maintainer"`
		WebsiteURL string `json:"websiteURL"`
		Email      string `json:"email,omitempty"`
		Help       struct {
			Online string `json:"online"`
		} `json:"help,omitempty"`
		Tools []Tool `json:"tools"`
	} `json:"packages"`
}

type Tool struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Systems []struct {
		Host            string `json:"host"`
		URL             string `json:"url"`
		ArchiveFileName string `json:"archiveFileName"`
		Checksum        string `json:"checksum"`
		Size            string `json:"size"`
	} `json:"systems"`
}
