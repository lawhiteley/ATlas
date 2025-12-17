package main

import (
	"ATlas/components"
	"fmt" // TODO: logging - slog?
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		index := components.Page("Luke", components.Atlas(), components.Panel(false, "law.png", "Luke"))

		index.Render(r.Context(), w)
	})

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", nil)
}
