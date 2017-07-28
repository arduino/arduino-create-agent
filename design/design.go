package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = API("arduino-create-agent", func() {
	Title("Arduino Create Agent")
	Description("A bridge from the user's computer and the Create platform")
	Host("localhost:9000")
	Scheme("http")
	BasePath("/")
	Consumes("application/json")
	Produces("application/json")

	Origin("*", func() {
		Methods("GET", "PUT", "POST", "DELETE")
		Headers("Authorization", "Origin", "X-Requested-With", "Content-Type", "Accept")
		Credentials()
	})
})
