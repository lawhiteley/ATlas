package main

import(
	"fmt" // TODO: logging - slog?
	"net/http"
	"ATlas/components"
)

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		index := components.Page("Luke", components.Atlas())

		index.Render(r.Context(), w)
	})

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", nil)
}
