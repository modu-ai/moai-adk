const cp = require('child_process');

function run(userInput) {
  cp.exec(userInput);
}

module.exports = { run };
