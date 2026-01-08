# Blockchain TechStack - 区块链技术栈

**版本**: v2.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang

---

## 目录

- [1. 区块链选择](#1-区块链选择)
- [2. 以太坊节点](#2-以太坊节点)
- [3. 智能合约开发](#3-智能合约开发)
- [4. Web3 库](#4-web3-库)
- [5. Flash Loan 协议](#5-flash-loan-协议)
- [6. MEV 工具](#6-mev-工具)

---

## 1. 区块链选择

### 1.1 Ethereum (以太坊)

**选择理由**：

1. **最大的 DEX 生态**
   - Uniswap（最大的 DEX）
   - SushiSwap
   - PancakeSwap

2. **成熟的 Flash Loan 协议**
   - Aave V3
   - Uniswap V3 Flash
   - Balancer

3. **活跃的 MEV 生态**
   - Flashbots
   - MEV-Boost
   - 丰富的 MEV 工具

4. **开发工具完善**
   - go-ethereum (Go 客户端)
   - Hardhat (开发框架)
   - OpenZeppelin (安全库)

### 1.2 网络选择

**主网 (Mainnet)**
- 生产环境
- 真实资金

**测试网 (Goerli)**
- 开发和测试
- 免费测试 ETH

**L2 (可选)**
- Arbitrum
- Optimism
- Polygon

---

## 2. 以太坊节点

### 2.1 节点类型

**Archive Node (归档节点)**

**用途**：
- 查询历史价格数据
- 监控 DEX 池子状态
- 查询历史交易

**部署方案**：

1. **本地部署**
   ```bash
   # 使用 Docker
   docker run -d \
     --name ethereum-node \
     -p 8545:8545 \
     -v /data/ethereum:/root/.ethereum \
     ethereum/client-go:latest \
     --http \
     --http.api=eth,net,web3 \
     --http.corsdomain="*"
   ```

2. **托管服务（推荐）**
   - Infura
   - Alchemy
   - QuickNode

### 2.2 节点连接

**go-ethereum**：

```go
import "github.com/ethereum/go-ethereum/ethclient"

client, err := ethclient.Dial("https://mainnet.infura.io/v3/YOUR_PROJECT_ID")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// 获取最新区块号
header, err := client.HeaderByNumber(context.Background(), nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println(header.Number.String())
```

### 2.3 WebSocket 订阅

```go
// 订阅新区块
headers := make(chan *types.Header)
sub, err := client.SubscribeNewHead(context.Background(), headers)
if err != nil {
    log.Fatal(err)
}

for {
    select {
    case err := <-sub.Err():
        log.Fatal(err)
    case header := <-headers:
        fmt.Println(header.Number.String())
    }
}
```

---

## 3. 智能合约开发

### 3.1 Solidity 版本

**Solidity 0.8.20+**

### 3.2 开发框架

**Hardhat**

**安装**：
```bash
npm install --save-dev hardhat
```

**项目初始化**：
```bash
npx hardhat init
```

**项目结构**：
```
contracts/
├── FlashLoanArbitrage.sol
├── interfaces/
│   ├── IFlashLoanReceiver.sol
│   └── IUniswapV3FlashCallback.sol
└── libraries/
    ├── ArbitrageLibrary.sol
    └── DexLibrary.sol

script/
└── deploy.js

test/
└── FlashLoanArbitrage.test.ts

hardhat.config.js
```

### 3.3 安全库

**OpenZeppelin**

```bash
npm install @openzeppelin/contracts
```

**使用**：
```solidity
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
```

### 3.4 编译和部署

**编译**：
```bash
npx hardhat compile
```

**部署**：
```javascript
// script/deploy.js
const hre = require("hardhat");

async function main() {
    const FlashLoanArbitrage = await hre.ethers.getContractFactory("FlashLoanArbitrage");
    const contract = await FlashLoanArbitrage.deploy();
    await contract.deployed();
    console.log("Deployed to:", contract.address);
}

main().catch((error) => {
    console.error(error);
    process.exitCode = 1;
});
```

**运行**：
```bash
npx hardhat run scripts/deploy.js --network goerli
```

---

## 4. Web3 库

### 4.1 go-ethereum

**版本**：v1.13+

**安装**：
```bash
go get github.com/ethereum/go-ethereum@latest
```

**核心功能**：

1. **区块查询**
```go
block, err := client.BlockByNumber(ctx, big.NewInt(1000))
```

2. **交易查询**
```go
tx, err := client.TransactionByHash(ctx, txHash)
```

3. **合约调用**
```go
contract := NewMyContract(contractAddress, client)
result, err := contract.MyMethod(nil)
```

4. **发送交易**
```go
auth, _ := bind.NewTransactorWithChainID(privateKey, chainID)
tx, err := contract.MyMethod(auth, arg1, arg2)
```

### 4.2 其他库

**abigen** - 生成 Go 绑定
```bash
abigen --sol=contracts/MyContract.sol --pkg=contracts --out=contracts/MyContract.go
```

---

## 5. Flash Loan 协议

### 5.1 Aave V3

**合约地址**：

- **主网**: `0x8787B6dF9124118DF7435e05DF517C7cA74d3B30`
- **Goerli**: `0x5041c7539368C6665DF49DA1c529b275D4A6b591`

**接口**：

```solidity
interface IFlashLoanReceiver {
    function executeOperation(
        address asset,
        uint256 amount,
        uint256 premium,
        address initiator,
        bytes calldata params
    ) external returns (bool);
}
```

**使用流程**：

1. 借款：`POOL.flashLoan(receiver, assets, amounts, interestRateModes, onBehalfOf, params, referralCode)`
2. 执行：回调 `executeOperation()`
3. 还款：在 `executeOperation()` 中还款

### 5.2 Uniswap V3 Flash

**合约地址**：

- **主网**: `0xE592427A0AEce92De3Edee1F18E0157C05861564`
- **Goerli**: 测试网地址

**接口**：

```solidity
interface IUniswapV3FlashCallback {
    function uniswapV3FlashCallback(
        uint256 fee0,
        uint256 fee1,
        bytes calldata data
    ) external;
}
```

### 5.3 Balancer

**合约地址**：

- **主网**: `0xBA12222222228d8Ba445958a75a0704d566BF2C8`

**接口**：

```solidity
interface IVaultFlashLoanReceiver {
    function receiveFlashLoan(
        IERC20[] memory tokens,
        uint256[] memory amounts,
        uint256[] memory feeAmounts,
        bytes memory userData
    ) external;
}
```

---

## 6. MEV 工具

### 6.1 Flashbots

**用途**：防止被抢跑，私密提交交易

**SDK**：

```bash
npm install @flashbots/ethers-provider-bundle
```

**使用**：

```javascript
const { FlashbotsBundleProvider } = require("@flashbots/ethers-provider-bundle");

const flashbotsProvider = await new FlashbotsBundleProvider(
    network.provider,
    network.signer,
    network.flashbotsRelay
);

const signedBundle = await flashbotsProvider.signBundle([
    transaction1,
    transaction2
]);

const simulation = await flashbotsProvider.simulate(signedBundle);
const submission = await flashbotsProvider.sendRawBundle(signedBundle);
```

### 6.2 MEV-Boost

**用途**：以太坊升级后的 MEV 提交流

**文档**：https://github.com/flashbots/mev-boost

### 6.3 Mempool 监控

**工具**：
- Etherscan API
- Mempool Explorer
- 自建节点

---

## 附录

### A. Gas 费优化

**EIP-1559**：

```go
// 建议的 Gas 价格
gasPrice, err := client.SuggestGasPrice(ctx)
gasTipCap, err := client.SuggestGasTipCap(ctx)

// 动态调整
fee, err := types.NewLondonFeeChainID(chainID)
tx.SetGasTipCap(gasTipCap)
tx.SetGasFeeCap(gasPrice)
```

### B. 滑点计算

```solidity
// Uniswap V2 滑点计算
uint256 amountOut = getAmountOut(amountIn, reserveIn, reserveOut);
uint256 slippage = amountOut * slippageTolerance / 1000;
uint256 minAmountOut = amountOut - slippage;
```

### C. 相关资源

- [Ethereum 开发者文档](https://ethereum.org/developers)
- [Hardhat 文档](https://hardhat.org/docs)
- [OpenZeppelin 合约](https://docs.openzeppelin.com/contracts)
- [go-ethereum 文档](https://geth.ethereum.org/docs)
- [Aave 开发者文档](https://docs.aave.com/developers)
- [Uniswap V3 文档](https://docs.uniswap.org/protocol/introduction)
- [Flashbots 文档](https://docs.flashbots.net/flashbots-auction/searchers/overview)

---

**相关文档**:
- [Backend_TechStack.md](./Backend_TechStack.md) - 后端技术栈
- [Modules/Flash_Loan_Contract.md](../Modules/Flash_Loan_Contract.md) - Flash Loan 合约设计
- [Modules/MEV_Engine.md](../Modules/MEV_Engine.md) - MEV 引擎设计
