// SPDX-License-Identifier: Apache-2.0

package main

type Reactions struct {
	PlusOne    int64  `json:"+1"`
	MinusOne   int64  `json:"-1"`
	Confused   int64  `json:"confused"`
	Eyes       int64  `json:"eyes"`
	Heart      int64  `json:"heart"`
	Hooray     int64  `json:"hooray"`
	Laugh      int64  `json:"laugh"`
	Rocket     int64  `json:"rocket"`
	TotalCount int64  `json:"total_count"`
	Url        string `json:"url"`
}

type Author struct {
	AvatarUrl         string `json:"avatar_url"`
	EventsUrl         string `json:"events_url"`
	FollowersUrl      string `json:"followers_url"`
	FollowingUrl      string `json:"following_url"`
	GistsUrl          string `json:"gists_url"`
	GravatarId        string `json:"gravatar_id"`
	HtmlUrl           string `json:"html_url"`
	Id                int64  `json:"id"`
	Login             string `json:"login"`
	NodeId            string `json:"node_id"`
	OrganizationsUrl  string `json:"organizations_url"`
	ReceivedEventsUrl string `json:"received_events_url"`
	ReposUrl          string `json:"repos_url"`
	SiteAdmin         bool   `json:"site_admin"`
	StarredUrl        string `json:"starred_url"`
	SubscriptionsUrl  string `json:"subscriptions_url"`
	Type              string `json:"type"`
	Url               string `json:"url"`
}

type Asset struct {
	BrowserDownloadUrl string `json:"browser_download_url"`
	ContentType        string `json:"content_type"`
	CreatedAt          string `json:"created_at"`
	DownloadCount      int64  `json:"download_count"`
	Id                 int64  `json:"id"`
	Label              string `json:"label"`
	Name               string `json:"name"`
	NodeId             string `json:"node_id"`
	Size               int64  `json:"size"`
	State              string `json:"state"`
	UpdatedAt          string `json:"updated_at"`
	Author             Author `json:"uploader"`
	Url                string `json:"url"`
}

type Release struct {
	Assets          []Asset   `json:"assets"`
	AssetsUrl       string    `json:"assets_url"`
	Author          Author    `json:"author"`
	Body            string    `json:"body"`
	CreatedAt       string    `json:"created_at"`
	Draft           bool      `json:"draft"`
	HtmlUrl         string    `json:"html_url"`
	Id              int64     `json:"id"`
	Name            string    `json:"name"`
	NodeId          string    `json:"node_id"`
	Prerelease      bool      `json:"prerelease"`
	PublishedAt     string    `json:"published_at"`
	Reactions       Reactions `json:"reactions"`
	TagName         string    `json:"tag_name"`
	TarballUrl      string    `json:"tarball_url"`
	TargetCommitish string    `json:"target_commitish"`
	UploadUrl       string    `json:"upload_url"`
	Url             string    `json:"url"`
	ZipballUrl      string    `json:"zipball_url"`
}
