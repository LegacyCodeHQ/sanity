package watch

import _ "embed"

//go:embed viewer.html
var indexHTML string

//go:embed viewer.js
var viewerJS string

//go:embed viewer_state.mjs
var viewerStateJS string

//go:embed viewer_protocol.mjs
var viewerProtocolJS string
