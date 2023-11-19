// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	noVersionAvailable              = "No new version available"
	youAreRunningAPreReleaseVersion = "You are running a pre-release version"
)

func TestCheckForNewRelease(t *testing.T) {

	client := createHTTPClient()
	version := "0.0.0"
	useragent := fmt.Sprintf("go-lambda-cleanup/%s", version)

	want := true
	got, msg, err := checkForNewRelease(client, version, useragent, url)
	if err != nil {
		t.Fatalf("Error checking for new release: %s", err)
	}
	if got != want && msg == noVersionAvailable || msg == youAreRunningAPreReleaseVersion {
		t.Fatalf("Error checking for new release: %s", err)
	}

	version = "100.0.0"
	want2 := true
	got2, msg2, err := checkForNewRelease(client, version, useragent, url)
	if err != nil {
		t.Fatalf("Error checking for new release: %s", err)
	}
	if got2 != want2 && msg2 != youAreRunningAPreReleaseVersion {
		t.Fatalf("Error checking for new release: %s", err)
	}
}

func TestErrorPath(t *testing.T) {

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Mock error response", http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	// Override the base URL to use the mock server URL
	baseURL := mockServer.URL

	client := createHTTPClient()
	version := "------"
	useragent := fmt.Sprintf("go-lambda-cleanup/%s", version)

	_, _, err := checkForNewRelease(client, version, useragent, baseURL)
	if err == nil {
		t.Fatalf("Error Expected: %s", err)
	}
}

func TestErrorPathJSON(t *testing.T) {

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("invalid-json-payload"))
		if err != nil {
			t.Fatalf("Error writing to response writer: %s", err)
		}
	}))
	defer mockServer.Close()

	// Override the base URL to use the mock server URL
	baseURL := mockServer.URL

	client := createHTTPClient()
	version := "------"
	useragent := fmt.Sprintf("go-lambda-cleanup/%s", version)

	_, _, err := checkForNewRelease(client, version, useragent, baseURL)
	if err == nil {
		t.Fatalf("Error Expected: %s", err)
	}
}

func TestErrorURL(t *testing.T) {

	client := createHTTPClient()
	version := "------"
	useragent := fmt.Sprintf("go-lambda-cleanup/%s", version)

	got, want, err := checkForNewRelease(client, version, useragent, "http://localhost:1234")
	if err == nil {
		t.Fatalf("Error Expected: %s", err)
	}
	if got != false && want != IssueMSG {
		t.Fatalf("Error Expected: %s", want)
	}
}

func TestCheckForNewReleaseNoNewRelease(t *testing.T) {

	var payload = []byte(`{
		"url": "https://api.github.com/repos/karl-cardenas-coding/go-lambda-cleanup/releases/127949448",
		"assets_url": "https://api.github.com/repos/karl-cardenas-coding/go-lambda-cleanup/releases/127949448/assets",
		"upload_url": "https://uploads.github.com/repos/karl-cardenas-coding/go-lambda-cleanup/releases/127949448/assets{?name,label}",
		"html_url": "https://github.com/karl-cardenas-coding/go-lambda-cleanup/releases/tag/v2.0.10",
		"id": 127949448,
		"author": {
		  "login": "github-actions[bot]",
		  "id": 41898282,
		  "node_id": "MDM6Qm90NDE4OTgyODI=",
		  "avatar_url": "https://avatars.githubusercontent.com/in/15368?v=4",
		  "gravatar_id": "",
		  "url": "https://api.github.com/users/github-actions%5Bbot%5D",
		  "html_url": "https://github.com/apps/github-actions",
		  "followers_url": "https://api.github.com/users/github-actions%5Bbot%5D/followers",
		  "following_url": "https://api.github.com/users/github-actions%5Bbot%5D/following{/other_user}",
		  "gists_url": "https://api.github.com/users/github-actions%5Bbot%5D/gists{/gist_id}",
		  "starred_url": "https://api.github.com/users/github-actions%5Bbot%5D/starred{/owner}{/repo}",
		  "subscriptions_url": "https://api.github.com/users/github-actions%5Bbot%5D/subscriptions",
		  "organizations_url": "https://api.github.com/users/github-actions%5Bbot%5D/orgs",
		  "repos_url": "https://api.github.com/users/github-actions%5Bbot%5D/repos",
		  "events_url": "https://api.github.com/users/github-actions%5Bbot%5D/events{/privacy}",
		  "received_events_url": "https://api.github.com/users/github-actions%5Bbot%5D/received_events",
		  "type": "Bot",
		  "site_admin": false
		},
		"node_id": "RE_kwDOFGhZlc4HoFqI",
		"tag_name": "v2.0.10",
		"target_commitish": "main",
		"name": "v2.0.10",
		"draft": false,
		"prerelease": false,
		"created_at": "2023-11-03T22:18:26Z",
		"published_at": "2023-11-03T22:18:37Z",
		"assets": [
		  {
			"url": "https://api.github.com/repos/karl-cardenas-coding/go-lambda-cleanup/releases/assets/133848132",
			"id": 133848132,
			"node_id": "RA_kwDOFGhZlc4H-lxE",
			"name": "go-lambda-cleanup-v2.0.10-darwin-amd64.zip",
			"label": "",
			"uploader": {
			  "login": "github-actions[bot]",
			  "id": 41898282,
			  "node_id": "MDM6Qm90NDE4OTgyODI=",
			  "avatar_url": "https://avatars.githubusercontent.com/in/15368?v=4",
			  "gravatar_id": "",
			  "url": "https://api.github.com/users/github-actions%5Bbot%5D",
			  "html_url": "https://github.com/apps/github-actions",
			  "followers_url": "https://api.github.com/users/github-actions%5Bbot%5D/followers",
			  "following_url": "https://api.github.com/users/github-actions%5Bbot%5D/following{/other_user}",
			  "gists_url": "https://api.github.com/users/github-actions%5Bbot%5D/gists{/gist_id}",
			  "starred_url": "https://api.github.com/users/github-actions%5Bbot%5D/starred{/owner}{/repo}",
			  "subscriptions_url": "https://api.github.com/users/github-actions%5Bbot%5D/subscriptions",
			  "organizations_url": "https://api.github.com/users/github-actions%5Bbot%5D/orgs",
			  "repos_url": "https://api.github.com/users/github-actions%5Bbot%5D/repos",
			  "events_url": "https://api.github.com/users/github-actions%5Bbot%5D/events{/privacy}",
			  "received_events_url": "https://api.github.com/users/github-actions%5Bbot%5D/received_events",
			  "type": "Bot",
			  "site_admin": false
			},
			"content_type": "application/zip",
			"state": "uploaded",
			"size": 8622699,
			"download_count": 1,
			"created_at": "2023-11-03T22:18:34Z",
			"updated_at": "2023-11-03T22:18:35Z",
			"browser_download_url": "https://github.com/karl-cardenas-coding/go-lambda-cleanup/releases/download/v2.0.10/go-lambda-cleanup-v2.0.10-darwin-amd64.zip"
		  },
		  {
			"url": "https://api.github.com/repos/karl-cardenas-coding/go-lambda-cleanup/releases/assets/133848129",
			"id": 133848129,
			"node_id": "RA_kwDOFGhZlc4H-lxB",
			"name": "go-lambda-cleanup-v2.0.10-darwin-arm64.zip",
			"label": "",
			"uploader": {
			  "login": "github-actions[bot]",
			  "id": 41898282,
			  "node_id": "MDM6Qm90NDE4OTgyODI=",
			  "avatar_url": "https://avatars.githubusercontent.com/in/15368?v=4",
			  "gravatar_id": "",
			  "url": "https://api.github.com/users/github-actions%5Bbot%5D",
			  "html_url": "https://github.com/apps/github-actions",
			  "followers_url": "https://api.github.com/users/github-actions%5Bbot%5D/followers",
			  "following_url": "https://api.github.com/users/github-actions%5Bbot%5D/following{/other_user}",
			  "gists_url": "https://api.github.com/users/github-actions%5Bbot%5D/gists{/gist_id}",
			  "starred_url": "https://api.github.com/users/github-actions%5Bbot%5D/starred{/owner}{/repo}",
			  "subscriptions_url": "https://api.github.com/users/github-actions%5Bbot%5D/subscriptions",
			  "organizations_url": "https://api.github.com/users/github-actions%5Bbot%5D/orgs",
			  "repos_url": "https://api.github.com/users/github-actions%5Bbot%5D/repos",
			  "events_url": "https://api.github.com/users/github-actions%5Bbot%5D/events{/privacy}",
			  "received_events_url": "https://api.github.com/users/github-actions%5Bbot%5D/received_events",
			  "type": "Bot",
			  "site_admin": false
			},
			"content_type": "application/zip",
			"state": "uploaded",
			"size": 8269741,
			"download_count": 2,
			"created_at": "2023-11-03T22:18:33Z",
			"updated_at": "2023-11-03T22:18:34Z",
			"browser_download_url": "https://github.com/karl-cardenas-coding/go-lambda-cleanup/releases/download/v2.0.10/go-lambda-cleanup-v2.0.10-darwin-arm64.zip"
		  },
		  {
			"url": "https://api.github.com/repos/karl-cardenas-coding/go-lambda-cleanup/releases/assets/133848128",
			"id": 133848128,
			"node_id": "RA_kwDOFGhZlc4H-lxA",
			"name": "go-lambda-cleanup-v2.0.10-linux-386.zip",
			"label": "",
			"uploader": {
			  "login": "github-actions[bot]",
			  "id": 41898282,
			  "node_id": "MDM6Qm90NDE4OTgyODI=",
			  "avatar_url": "https://avatars.githubusercontent.com/in/15368?v=4",
			  "gravatar_id": "",
			  "url": "https://api.github.com/users/github-actions%5Bbot%5D",
			  "html_url": "https://github.com/apps/github-actions",
			  "followers_url": "https://api.github.com/users/github-actions%5Bbot%5D/followers",
			  "following_url": "https://api.github.com/users/github-actions%5Bbot%5D/following{/other_user}",
			  "gists_url": "https://api.github.com/users/github-actions%5Bbot%5D/gists{/gist_id}",
			  "starred_url": "https://api.github.com/users/github-actions%5Bbot%5D/starred{/owner}{/repo}",
			  "subscriptions_url": "https://api.github.com/users/github-actions%5Bbot%5D/subscriptions",
			  "organizations_url": "https://api.github.com/users/github-actions%5Bbot%5D/orgs",
			  "repos_url": "https://api.github.com/users/github-actions%5Bbot%5D/repos",
			  "events_url": "https://api.github.com/users/github-actions%5Bbot%5D/events{/privacy}",
			  "received_events_url": "https://api.github.com/users/github-actions%5Bbot%5D/received_events",
			  "type": "Bot",
			  "site_admin": false
			},
			"content_type": "application/zip",
			"state": "uploaded",
			"size": 8169046,
			"download_count": 2,
			"created_at": "2023-11-03T22:18:32Z",
			"updated_at": "2023-11-03T22:18:33Z",
			"browser_download_url": "https://github.com/karl-cardenas-coding/go-lambda-cleanup/releases/download/v2.0.10/go-lambda-cleanup-v2.0.10-linux-386.zip"
		  },
		  {
			"url": "https://api.github.com/repos/karl-cardenas-coding/go-lambda-cleanup/releases/assets/133848138",
			"id": 133848138,
			"node_id": "RA_kwDOFGhZlc4H-lxK",
			"name": "go-lambda-cleanup-v2.0.10-linux-amd64.zip",
			"label": "",
			"uploader": {
			  "login": "github-actions[bot]",
			  "id": 41898282,
			  "node_id": "MDM6Qm90NDE4OTgyODI=",
			  "avatar_url": "https://avatars.githubusercontent.com/in/15368?v=4",
			  "gravatar_id": "",
			  "url": "https://api.github.com/users/github-actions%5Bbot%5D",
			  "html_url": "https://github.com/apps/github-actions",
			  "followers_url": "https://api.github.com/users/github-actions%5Bbot%5D/followers",
			  "following_url": "https://api.github.com/users/github-actions%5Bbot%5D/following{/other_user}",
			  "gists_url": "https://api.github.com/users/github-actions%5Bbot%5D/gists{/gist_id}",
			  "starred_url": "https://api.github.com/users/github-actions%5Bbot%5D/starred{/owner}{/repo}",
			  "subscriptions_url": "https://api.github.com/users/github-actions%5Bbot%5D/subscriptions",
			  "organizations_url": "https://api.github.com/users/github-actions%5Bbot%5D/orgs",
			  "repos_url": "https://api.github.com/users/github-actions%5Bbot%5D/repos",
			  "events_url": "https://api.github.com/users/github-actions%5Bbot%5D/events{/privacy}",
			  "received_events_url": "https://api.github.com/users/github-actions%5Bbot%5D/received_events",
			  "type": "Bot",
			  "site_admin": false
			},
			"content_type": "application/zip",
			"state": "uploaded",
			"size": 8631978,
			"download_count": 15,
			"created_at": "2023-11-03T22:18:35Z",
			"updated_at": "2023-11-03T22:18:36Z",
			"browser_download_url": "https://github.com/karl-cardenas-coding/go-lambda-cleanup/releases/download/v2.0.10/go-lambda-cleanup-v2.0.10-linux-amd64.zip"
		  },
		  {
			"url": "https://api.github.com/repos/karl-cardenas-coding/go-lambda-cleanup/releases/assets/133848139",
			"id": 133848139,
			"node_id": "RA_kwDOFGhZlc4H-lxL",
			"name": "go-lambda-cleanup-v2.0.10-windows-amd64.zip",
			"label": "",
			"uploader": {
			  "login": "github-actions[bot]",
			  "id": 41898282,
			  "node_id": "MDM6Qm90NDE4OTgyODI=",
			  "avatar_url": "https://avatars.githubusercontent.com/in/15368?v=4",
			  "gravatar_id": "",
			  "url": "https://api.github.com/users/github-actions%5Bbot%5D",
			  "html_url": "https://github.com/apps/github-actions",
			  "followers_url": "https://api.github.com/users/github-actions%5Bbot%5D/followers",
			  "following_url": "https://api.github.com/users/github-actions%5Bbot%5D/following{/other_user}",
			  "gists_url": "https://api.github.com/users/github-actions%5Bbot%5D/gists{/gist_id}",
			  "starred_url": "https://api.github.com/users/github-actions%5Bbot%5D/starred{/owner}{/repo}",
			  "subscriptions_url": "https://api.github.com/users/github-actions%5Bbot%5D/subscriptions",
			  "organizations_url": "https://api.github.com/users/github-actions%5Bbot%5D/orgs",
			  "repos_url": "https://api.github.com/users/github-actions%5Bbot%5D/repos",
			  "events_url": "https://api.github.com/users/github-actions%5Bbot%5D/events{/privacy}",
			  "received_events_url": "https://api.github.com/users/github-actions%5Bbot%5D/received_events",
			  "type": "Bot",
			  "site_admin": false
			},
			"content_type": "application/zip",
			"state": "uploaded",
			"size": 8759318,
			"download_count": 1,
			"created_at": "2023-11-03T22:18:36Z",
			"updated_at": "2023-11-03T22:18:37Z",
			"browser_download_url": "https://github.com/karl-cardenas-coding/go-lambda-cleanup/releases/download/v2.0.10/go-lambda-cleanup-v2.0.10-windows-amd64.zip"
		  }
		],
		"tarball_url": "https://api.github.com/repos/karl-cardenas-coding/go-lambda-cleanup/tarball/v2.0.10",
		"zipball_url": "https://api.github.com/repos/karl-cardenas-coding/go-lambda-cleanup/zipball/v2.0.10",
		"body": "## [2.0.10](https://github.com/karl-cardenas-coding/go-lambda-cleanup/compare/v2.0.9...v2.0.10) (2023-11-03)\n\n\n### Bug Fixes\n\n* updated dependencies ([778b9e2](https://github.com/karl-cardenas-coding/go-lambda-cleanup/commit/778b9e2a457544c1874582abdca089d2123578a1))\n\n\n\n"
	  }`)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Write an invalid JSON payload to the response writer
		_, err := w.Write(payload)
		if err != nil {
			t.Fatalf("Error writing to response writer: %s", err)
		}
	}))
	defer mockServer.Close()

	// Override the base URL to use the mock server URL
	baseURL := mockServer.URL

	client := createHTTPClient()
	version := "2.0.10"
	useragent := fmt.Sprintf("go-lambda-cleanup/%s", version)

	want := false
	got, msg, err := checkForNewRelease(client, version, useragent, baseURL)
	if err != nil {
		t.Fatalf("Error encountered: %s", err)
	}

	if got != want && msg != noVersionAvailable {
		t.Fatalf("Error checking for new release: %s", msg)
	}
}
