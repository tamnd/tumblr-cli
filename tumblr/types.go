package tumblr

// Post represents a Tumblr post.
type Post struct {
	Rank     int    `json:"rank"      csv:"rank"      tsv:"rank"`
	ID       string `json:"id"        csv:"id"        tsv:"id"`
	Type     string `json:"type"      csv:"type"      tsv:"type"`
	Blog     string `json:"blog"      csv:"blog"      tsv:"blog"`
	Summary  string `json:"summary"   csv:"summary"   tsv:"summary"`
	Date     string `json:"date"      csv:"date"      tsv:"date"`
	Notes    int    `json:"notes"     csv:"notes"     tsv:"notes"`
	Tags     string `json:"tags"      csv:"tags"      tsv:"tags"`
	URL      string `json:"url"       csv:"url"       tsv:"url"`
}

// Blog represents a Tumblr blog.
type Blog struct {
	Name        string `json:"name"        csv:"name"        tsv:"name"`
	Title       string `json:"title"       csv:"title"       tsv:"title"`
	Description string `json:"description" csv:"description" tsv:"description"`
	Posts       int    `json:"posts"       csv:"posts"       tsv:"posts"`
	Updated     int64  `json:"updated"     csv:"updated"     tsv:"updated"`
	URL         string `json:"url"         csv:"url"         tsv:"url"`
}
