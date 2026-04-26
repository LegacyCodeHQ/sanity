package formatters

import (
	"path/filepath"
	"sort"
)

var extensionColorPalette = []string{
	"lightblue", "lightyellow", "mistyrose", "lightsalmon",
	"lightpink", "lavender", "peachpuff", "plum", "powderblue", "khaki",
	"palegoldenrod", "thistle",
}

func getExtensionColors(fileNames []string) map[string]string {
	uniqueExtensions := make(map[string]bool)
	for _, fileName := range fileNames {
		ext := filepath.Ext(fileName)
		if ext != "" {
			uniqueExtensions[ext] = true
		}
	}

	sortedExtensions := make([]string, 0, len(uniqueExtensions))
	for ext := range uniqueExtensions {
		sortedExtensions = append(sortedExtensions, ext)
	}
	sort.Strings(sortedExtensions)

	extensionColors := make(map[string]string)
	for i, ext := range sortedExtensions {
		color := extensionColorPalette[i%len(extensionColorPalette)]
		extensionColors[ext] = color
	}

	return extensionColors
}
