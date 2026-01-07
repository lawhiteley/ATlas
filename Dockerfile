FROM golang:1.25-alpine

RUN apk add --no-cache gcc musl-dev nodejs npm
RUN npm install -g bun

WORKDIR /app
COPY . .

RUN bun install
RUN bunx tailwindcss -i ./assets/input.css -o ./static/css/output.css --minify
RUN go tool templ generate
RUN go build -o main .

EXPOSE 8080
CMD ["./main"]