// Package web holds the static front-end files that are compiled into the jot
// binary, so generated sites do not depend on the working directory.
package web

import "embed"

// Assets contains the CSS and JavaScript copied into every generated site,
// rooted at templates/assets. Embedding the directory rather than a glob
// leaves out dotfiles such as .DS_Store.
//
//go:embed templates/assets
var Assets embed.FS

// AssetsDir is the directory within Assets that holds the asset files.
const AssetsDir = "templates/assets"
