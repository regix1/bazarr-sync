package bazarr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"

	"github.com/pterm/pterm"
	"github.com/regix1/bazarr-sync/internal/client"
	"github.com/regix1/bazarr-sync/internal/config"
)

func QueryMovies(cfg config.Config) (movies_info, error) {
	c := client.GetClient(cfg.ApiToken)
	url, _ := url.JoinPath(cfg.ApiUrl, "movies")
	resp, err := c.Get(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Connection Error: ", err)
		return movies_info{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Fprintln(os.Stderr, "Connection Error: Response status is not 200. Are you sure the address/port are correct?")
		return movies_info{}, errors.New("Error: Status code not 200")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Reading Url Error: ", err)
		return movies_info{}, err
	}
	var data movies_info
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error in Unmarshaling json body", err)
		return movies_info{}, err
	}
	return data, nil
}

func QuerySeries(cfg config.Config) (shows_info, error) {
	c := client.GetClient(cfg.ApiToken)
	url, _ := url.JoinPath(cfg.ApiUrl, "series")
	resp, err := c.Get(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Connection Error: ", err)
		return shows_info{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Fprintln(os.Stderr, "Connection Error: Response status is not 200. Are you sure the address/port are correct?")
		return shows_info{}, errors.New("Error: Status code not 200")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Reading Url Error: ", err)
		return shows_info{}, err
	}
	var data shows_info
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error in Unmarshaling json body", err)
		return shows_info{}, err
	}
	return data, nil
}

func QueryEpisodes(cfg config.Config, seriesId int) (episodes_info, error) {
	c := client.GetClient(cfg.ApiToken)
	u, _ := url.JoinPath(cfg.ApiUrl, "episodes")
	_url, _ := url.Parse(u)
	queryUrl := _url.Query()
	queryUrl.Set("seriesid[]", strconv.Itoa(seriesId))
	_url.RawQuery = queryUrl.Encode()

	resp, err := c.Get(_url.String())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Connection Error: ", err)
		return episodes_info{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Fprintln(os.Stderr, "Connection Error: Response status is not 200. Are you sure the address/port are correct?")
		return episodes_info{}, errors.New("Error: Status code not 200")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Reading Url Error: ", err)
		return episodes_info{}, err
	}
	var data episodes_info
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error in Unmarshaling json body", err)
		return episodes_info{}, err
	}
	return data, nil
}

func GetSyncParams(_type string, id int, subtitleInfo subtitle_info) Sync_params {
	var params Sync_params
	params.Action = "sync"
	params.Path = subtitleInfo.Path
	params.Id = id
	params.Lang = subtitleInfo.Code2
	params.Type = _type
	params.Gss = "False"
	params.No_framerate_fix = "False"
	return params
}

// Updated Sync function with detailed error reporting
func Sync(cfg config.Config, params Sync_params) (bool, string) {
	c := client.GetClient(cfg.ApiToken)
	u, _ := url.JoinPath(cfg.ApiUrl, "subtitles")

	_url, _ := url.Parse(u)
	queryUrl := _url.Query()
	queryUrl.Set("path", params.Path)
	queryUrl.Set("id", strconv.Itoa(params.Id))
	queryUrl.Set("action", "sync")
	queryUrl.Set("language", params.Lang)
	queryUrl.Set("type", params.Type)
	_url.RawQuery = queryUrl.Encode()

	resp, err := c.Patch(_url.String())
	if err != nil {
		return false, fmt.Sprintf("Connection error: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Different status codes mean different things
	switch resp.StatusCode {
	case 204:
		// Success
		return true, "Success"
	case 304:
		// Not modified - already synced
		return false, "Already synced (no changes needed)"
	case 400:
		// Bad request
		if len(body) > 0 {
			return false, fmt.Sprintf("Bad request: %s", string(body))
		}
		return false, "Bad request (check subtitle file)"
	case 404:
		// Not found
		return false, "Subtitle file not found"
	case 409:
		// Conflict - usually means already in sync
		return false, "Already in perfect sync"
	case 500:
		// Server error
		if len(body) > 0 {
			// Check for specific error messages
			bodyStr := string(body)
			if contains(bodyStr, "already synchronized") || contains(bodyStr, "already in sync") {
				return false, "Already synchronized"
			}
			if contains(bodyStr, "subsync") || contains(bodyStr, "ffmpeg") {
				return false, "Sync tool not available (check subsync/ffmpeg)"
			}
			return false, fmt.Sprintf("Server error: %s", bodyStr)
		}
		return false, "Server error during sync"
	default:
		if len(body) > 0 {
			return false, fmt.Sprintf("Status %d: %s", resp.StatusCode, string(body))
		}
		return false, fmt.Sprintf("Unknown status: %d", resp.StatusCode)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}

func HealthCheck(cfg config.Config) {
	c := client.GetClient(cfg.ApiToken)
	url, _ := url.JoinPath(cfg.ApiUrl, "system/status")
	resp, err := c.Get(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Connection Error: ", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Fprintln(os.Stderr, "Connection Error: Response status is not 200. Are you sure the address/port are correct?")
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Reading Url Error: ", err)
		return
	}
	var data version
	json.Unmarshal(body, &data)
	fmt.Println("Bazarr version: ", pterm.LightBlue(data.Data.Bazarr_version))
}
