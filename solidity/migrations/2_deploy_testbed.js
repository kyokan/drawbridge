var TestToken = artifacts.require('./TestToken.sol');
var Lightning = artifacts.require('./Lightning.sol');

module.exports = function(deployer) {
  deployer.deploy(TestToken)
    .then((instance) => deployer.deploy(Lightning, instance.address))
};
