package diskplayer

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/zmb3/spotify"
	"os"
	"net/http"
	"time"
)

// PlayPath will play an album or playlist by reading a Spotify URI from a file whose filepath is passed into the
// function.
// An error is returned if one is encountered.
func PlayPath(c Client, p string) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()

	// TODO: scan them into different strings and call webhook with "event" string
	s := bufio.NewScanner(f)
	var l string
	var lightColor string
	if s.Scan() {
		l = s.Text()
	}
	if s.Scan() {
		lightColor = s.Text()
	}

	if l == "" {
		return fmt.Errorf("unable to read line from path: %s", p)
	}

	if lightColor != "" {
		//SetLight(lightColor)
	}

	return PlayUri(c, l)
}

// Set my smartbulb using ifttt and maker webhooks
func SetLight(color string) error {
	url := fmt.Sprintf("https://maker.ifttt.com/trigger/%s/with/key/dvzmHn0VCQpK3mrPhZ0O96", color)

	tr := &http.Transport{
		MaxIdleConns:       2,
  		IdleConnTimeout:    3 * time.Second,
  		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
        resp, err := client.Get(url)
	if err != nil {
		return err
	}
	fmt.Println(resp.Body)
	resp.Body.Close()
	return nil
}

// PlayURI will play the album or playlist Spotify URI that is passed in to the function.
// An error is returned if one is encountered.
func PlayUri(c Client, u string) error {
	if u == "" {
		return errors.New("spotify URI is required")
	}

	spotifyUri := spotify.URI(u)

	n := ConfigValue(SPOTIFY_DEVICE_NAME)
	ds, err := c.PlayerDevices()
	if err != nil {
		return err
	}

	activeID := activePlayerId(&ds)

	playerID := diskplayerId(&ds, n)
	if playerID == "" {
		return fmt.Errorf("client identified by %s not found", n)
	}

	if activeID != "" && activeID != playerID {
		err := c.Pause()
		if err != nil {
			return err
		}
		err = c.TransferPlayback(playerID, false)
		if err != nil {
			return err
		}
	}

	o := &spotify.PlayOptions{
		DeviceID:        &playerID,
		PlaybackContext: &spotifyUri,
	}

	c.Shuffle(true)
	return c.PlayOpt(o)
}

// Pause will pause the Spotify playback if the Diskplayer is the currently active Spotify device.
// An error is returned if one is encountered.
func Pause(c Client) error {
	n := ConfigValue(SPOTIFY_DEVICE_NAME)
	ds, err := c.PlayerDevices()
	if err != nil {
		return err
	}

	activeID := activePlayerId(&ds)
	if activeID == "" {
		return nil
	}

	playerID := diskplayerId(&ds, n)
	if playerID == "" {
		return fmt.Errorf("client identified by %s not found", n)
	}

	if activeID == playerID {
		err := c.Pause()
		if err != nil {
			return err
		}
	}

	return nil
}

// activePlayerIds iterates through the provided player devices and returns the active ID. If there is no active
// Spotify client device the ID will be returned as a nil pointer.
func activePlayerId(ds *[]spotify.PlayerDevice) spotify.ID {
	for _, d := range *ds {
		if d.Active {
			return d.ID
		}
	}

	return ""
}

// diskplayerId returns the Spotify ID for the Spotify client whose name is provided in the parameter list,
// or a nil pointer if no matching device is found.
func diskplayerId(ds *[]spotify.PlayerDevice, n string) spotify.ID {
	for _, d := range *ds {
		if d.Name == n {
			return d.ID
		}
	}

	return ""
}
