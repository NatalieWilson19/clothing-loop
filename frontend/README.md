# Vite Frontend

## Development install

1. Install NodeJS, this project tries to stay on the latest LTS version (v18 as of writing)

   For Windows or Mac: https://nodejs.org/en/
   For Linux: https://nodejs.org/en/download/package-manager/

2. `cd` into the frontend directory.
3. Copy the env file `cp .env-example .env`.
4. Run `npm i` to install all necessary dependencies.
5. Ensure that the api server is running.
6. Run `npm run start`

## Commands

In the project directory, you can run:

### Run the app in the development mode

The page will reload if you make edits.
You will also see any errors in the console.

`npm run start`

Then open [http://localhost:3000](http://localhost:3000) on your preferred browser.

### Test the app on your phone

1. Run `npm run start:lan`
2. Open port `3000` on your pc
3. Figure out what LAN IP your pc is at generally starts with `192.168.x.x`
4. Open your pc LAN IP on your phone e.g. `http://192.168.1.68:3000/`

### Build the app for production to the `build` folder

It correctly bundles React in production mode and optimizes the build for the best performance.

The build is minified and the filenames include the hashes.
Your app is ready to be deployed!

`npm run build`

### Lint

[Prettier](https://prettier.io/) is used as linter.
Please install the prettier addon to your preferred editor and ensure that you lint before you save or commit.

Tests using Prettier if the code is styled correctly.

`npm run lint:test`

Tries to fix any missing spaces or tabs required by Prettier

`npm run lint`

## Libraries

### TailwindCSS

Motive behind moving to tailwind is to decrease the bundle size of the website [Tailwind CSS vs. CSS: An in-depth comparison (speed, file size, etc.)](https://www.programonaut.com/tailwind-css-vs-css-an-in-depth-comparison-speed-file-size-etc/)

### Flags

Flag icons come from: https://flagicons.lipis.dev/
[MIT LICENSE](https://github.com/lipis/flag-icons/blob/main/LICENSE)

### Icons

Icons on this website uses https://feathericons.com/
[MIT LICENSE](https://github.com/feathericons/feather/blob/master/LICENSE)
