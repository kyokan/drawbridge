var TestToken = artifacts.require('./TestToken.sol');
var UTXOToken = artifacts.require('./UTXOToken.sol');

module.exports = function(deployer) {
  deployer.deploy(TestToken)
    .then((instance) => {
      if (process.env.NODE_ENV === 'development') {
        console.log('  Minting 1 million tokens for account 0.');

        return instance.mint('0x627306090abab3a6e1400e9345bc60c78a8bef57', 1000000)
          .then(() => instance)
      }

      return instance;
    })
    .then((instance) => deployer.deploy(UTXOToken, instance.address))
};
