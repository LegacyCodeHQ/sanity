package watch

import _ "embed"

//go:embed viewer.html
var indexHTML string

//go:embed viewer.js
var viewerJS string
