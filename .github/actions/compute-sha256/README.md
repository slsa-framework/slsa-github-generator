# Compute SHA256

## How to build this action in development
- Install node
- Install typescript
- `npm install`
- `npm run all`

## Compare the dist/index.js with the actual source
This is to ensure that the source is correct.
- `diff dist/index.js prdist/index.js`. In this case the prdist/index.js is the Pull Request version.
- The diff should be empty

