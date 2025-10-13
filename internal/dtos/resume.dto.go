package dtos

type Resume struct {
	Header         Header       `json:"header"`
	ProfileSummary string       `json:"profileSummary,omitempty"`
	Skills         []Skills     `json:"skills"`
	Experiences    []Experience `json:"experiences"`
	Education      []Education  `json:"education"`
	Projects       []Project    `json:"projects"`
	Awards         []Award      `json:"awards,omitempty"`
	SectionOrder   []string     `json:"sectionOrder"`
}

type Header struct {
	Fullname    string `json:"fullname"`
	JobTitle    string `json:"jobTitle"`
	Location    string `json:"location,omitempty"`
	Email       string `json:"email"`
	Phone       string `json:"phone,omitempty"`
	LinkedIn    string `json:"linkedin,omitempty"`
	LinkedInURL string `json:"linkedinUrl,omitempty"`
	Github      string `json:"github,omitempty"`
	GithubURL   string `json:"githubUrl,omitempty"`
	Website     string `json:"website,omitempty"`
	WebsiteURL  string `json:"websiteUrl,omitempty"`
}

type Skills struct {
	Title  string   `json:"title"`
	Values []string `json:"values"`
}

type Experience struct {
	Company      string   `json:"company"`
	Occupation   string   `json:"occupation"`
	StartDate    string   `json:"startDate"`
	EndDate      string   `json:"endDate"`
	Location     string   `json:"location"`
	Descriptions []string `json:"desc"`
}

type Education struct {
	Degree       string   `json:"degree"`
	Institution  string   `json:"institution"`
	StartDate    string   `json:"startDate"`
	EndDate      string   `json:"endDate"`
	Location     string   `json:"location"`
	Descriptions []string `json:"desc"`
}

type Project struct {
	Title        string   `json:"title"`
	Link         string   `json:"link"`
	Subtitle     string   `json:"subtitle"`
	Descriptions []string `json:"desc"`
}

type Award struct {
	Title        string   `json:"title"`
	Link         string   `json:"link"`
	Issuer       string   `json:"issuer"`
	Date         string   `json:"date"`
	Descriptions []string `json:"desc"`
}
