package bazarr

type version struct {
	Data struct {
		Bazarr_version string `json:"bazarr_version"`
	} `json:"data"`
}

type Sync_params struct {
	Path             string `json:"path"`
	Id               int    `json:"id"`
	Action           string `json:"action"`
	Lang             string `json:"language"`
	Type             string `json:"type"`
	Gss              string `json:"gss"`
	No_framerate_fix string `json:"no_fix_framerate"`
}

type movies_info struct {
	Data []movie `json:"data"`
}

type shows_info struct {
	Data []struct {
		Title          string `json:"title"`
		Monitored      bool   `json:"monitored"`
		SonarrSeriesId int    `json:"sonarrSeriesId"`
		ImdbId         string `json:"imdbId"`
	} `json:"data"`
}

type episodes_info struct {
	Data []episode `json:"data"`
}

type episode struct {
	Title           string          `json:"title"`
	Monitored       bool            `json:"monitored"`
	SonarrEpisodeId int             `json:"sonarrEpisodeId"`
	Subtitles       []subtitle_info `json:"subtitles"`
}

type movie struct {
	Title     string          `json:"title"`
	Monitored bool            `json:"monitored"`
	RadarrId  int             `json:"radarrId"`
	ImdbId    string          `json:"imdbId"`
	Subtitles []subtitle_info `json:"subtitles"`
}

type subtitle_info struct {
	Path      string `json:"path"`
	Code2     string `json:"code2"`
	File_size int    `json:"file_size"`
}
