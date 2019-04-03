package design

import . "goa.design/goa/dsl"

var _ = Service("docs", func() {
	HTTP(func() {
		Path("/docs")
	})
	Files("/pkgs", "docs/pkgs.html")
})
