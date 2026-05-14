// SPDX-License-Identifier: Apache-2.0

package main

type RepositoryInfo struct {
	Description string `json:"description"`
}

type Asset struct {
	Digest string `json:"digest"`
	Name   string `json:"name"`
	Url    string `json:"url"`
}

type Release struct {
	Assets  []Asset `json:"assets"`
	Name    string  `json:"name"`
	TagName string  `json:"tag_name"`
}
