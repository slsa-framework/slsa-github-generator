### Links
https://stackoverflow.com/questions/15806241/how-to-specify-local-modules-as-npm-package-dependencies
https://github.com/bersling/typescript-library-starter/blob/master
https://www.tsmean.com/articles/how-to-write-a-typescript-library/

### Error and fix
To avoid `npm ci` error:
```shell
npm ERR! code EUSAGE
npm ERR! 
npm ERR! `npm ci` can only install packages when your package.json and package-lock.json or npm-shrinkwrap.json are in sync. Please update your lock file with `npm install` before continuing.
npm ERR! 
npm ERR! Missing: tscommon@0.0.0 from lock file
npm ERR! 
npm ERR! Clean install a project
```

Use:
```shell
$ cd tscommon
$ npm pack
$ cd ../another-action
$ npm install --save ../tscommon/tscommon-0.0.0.tgz
```

### Changes
Any changes to this code need to be reflected in dependent Actions as shown above.
Run the script:
```bash
$ cd tscommon
$ npm run all
$ npm run package
$ bash update-actions.sh
```