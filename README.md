# OpenUEM - Console

Repository containing the OpenUEM web console used to manage endpoints

## Development

We use Air to re-run go build after changes to templates: `air`

We must generate go files every time we modify a view created by templ. We watch for new changes using: `templ generate --watch`

We use npm to generate tailwind css: `npm run watch-css`

We use esbuild to create a JS bundle for the project: `npm run build`

## Woodpecker

This repository has been connected with Woodpecker CI thanks to the file .woodpecker.yml

A Cloudflare tunnel is used to run CI pipelines in my local environment
