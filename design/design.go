package design

import . "goa.design/goa/dsl"

var _ = API("arduino-create-agent", func() {
	Title("Arduino Create Agent")
	Description(`A companion of Arduino Create. 
	Allows the website to perform operations on the user computer, 
	such as detecting which boards are connected and upload sketches on them.`)
	HTTP(func() {
		Path("/v2")
	})
})
