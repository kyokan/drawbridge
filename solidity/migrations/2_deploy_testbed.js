var TestToken = artifacts.require('./TestToken.sol');
var LightningERC20 = artifacts.require('./LightningERC20.sol');

module.exports = function(deployer) {
  deployer.deploy(TestToken)
    .then((instance) => deployer.deploy(LightningERC20, instance.address))
};
