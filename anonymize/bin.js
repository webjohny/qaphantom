const anonymize = require('.') /* the current working directory so that means main.js because of package.json */
let proxy = process.argv[2] /* what the user enters as first argument */

console.log(
    anonymize(proxy)
)