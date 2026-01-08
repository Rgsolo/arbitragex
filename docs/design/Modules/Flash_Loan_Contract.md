# Flash Loan Contract - Flash Loan 合约

**版本**: v1.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang
**优先级**: ⭐⭐⭐⭐⭐

---

## 目录

- [1. 模块概述](#1-模块概述)
- [2. Flash Loan 原理](#2-flash-loan-原理)
- [3. 智能合约设计](#3-智能合约设计)
- [4. Aave 集成](#4-aave-集成)
- [5. Uniswap V3 Flash](#5-uniswap-v3-flash)
- [6. Balancer Flash](#6-balancer-flash)
- [7. Solidity 实现](#7-solidity-实现)
- [8. 安全考虑](#8-安全考虑)
- [9. 部署和测试](#9-部署和测试)
- [10. Gas 优化](#10-gas-优化)

---

## 1. 模块概述

### 1.1 Flash Loan 简介

Flash Loan（闪电贷）是一种无需抵押的即时借贷方式，借款和还款必须在同一个交易区块内完成。

**特点**：
- 无需抵押
- 零信用风险
- 必须在同一区块内归还
- 支付少量手续费（0.09%）

**用途**：
- 套利交易
- 清算抵押品
- 更换抵押品
- 自我清算

### 1.2 主要协议

| 协议 | 手续费 | 链 | 特点 |
|------|--------|-----|------|
| Aave V3 | 0.05% | ETH, MATIC, ARB | 支持多资产 |
| Uniswap V3 | 0.3% | ETH | DEX 原生 |
| Balancer | 0% | ETH, MATIC | 无手续费 |

### 1.3 技术选型

| 技术栈 | 版本 | 用途 |
|--------|------|------|
| Solidity | 0.8.20+ | 智能合约开发 |
| Hardhat | latest | 开发框架 |
| OpenZeppelin | ^5.0 | 安全库 |
| go-ethereum | v1.13+ | 合约交互 |

---

## 2. Flash Loan 原理

### 2.1 工作流程

```
┌─────────────────────────────────────────────────────────┐
│                   Flash Loan 流程                        │
└─────────────────────────────────────────────────────────┘

1. 借款请求
   ├─ 用户向 Flash Loan 提供商发起借款
   └─ 指定借款金额和回调函数

2. 资金转账
   ├─ 提供商将资金转给借款人
   └─ 触发借款人的回调函数

3. 执行业务逻辑
   ├─ 套利交易
   ├─ 清算操作
   └─ 其他自定义逻辑

4. 还款验证
   ├─ 检查资金是否归还
   ├─ 检查手续费是否支付
   └─ 如果验证失败，整个交易回滚

5. 交易完成
   └─ 区块链状态更新
```

### 2.2 核心约束

**原子性**：整个操作必须在同一交易中完成

```solidity
// 伪代码
function executeFlashLoan(uint256 amount) {
    uint256 balanceBefore = address(this).balance;

    // 1. 借款
    lendingPool.flashLoan(amount);

    // 2. 执行业务逻辑
    doArbitrage(amount);

    // 3. 还款
    uint256 balanceAfter = address(this).balance;
    require(balanceAfter >= balanceBefore + fee, "Repayment failed");
}
```

---

## 3. 智能合约设计

### 3.1 合约架构

```
contracts/
├── FlashLoanArbitrage.sol          # Flash Loan 套利合约
├── interfaces/
│   ├── IFlashLoanReceiver.sol      # Aave 回调接口
│   ├── IUniswapV3FlashCallback.sol # Uniswap V3 回调接口
│   └── IVaultFlashLoanReceiver.sol # Balancer 回调接口
└── libraries/
    ├── ArbitrageLibrary.sol        # 套利逻辑库
    └── DexLibrary.sol              # DEX 操作库
```

### 3.2 核心接口

#### 3.2.1 Aave 回调接口

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

interface IFlashLoanReceiver {
    /**
     * @dev 执行 Flash Loan 回调
     * @param asset 借款资产地址
     * @param amount 借款金额
     * @param premium 手续费
     * @param initiator 发起人地址
     * @param params 参数（编码）
     * @return 成功返回 true
     */
    function executeOperation(
        address asset,
        uint256 amount,
        uint256 premium,
        address initiator,
        bytes calldata params
    ) external returns (bool);
}
```

#### 3.2.2 Uniswap V3 回调接口

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

interface IUniswapV3FlashCallback {
    /**
     * @dev 执行 Uniswap V3 Flash 回调
     * @param fee0 token0 手续费
     * @param fee1 token1 手续费
     * @param data 参数（编码）
     */
    function uniswapV3FlashCallback(
        uint256 fee0,
        uint256 fee1,
        bytes calldata data
    ) external;
}
```

#### 3.2.3 Balancer 回调接口

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

interface IVaultFlashLoanReceiver {
    /**
     * @dev 执行 Balancer Flash 回调
     * @param tokens 借款代币地址数组
     * @param amounts 借款金额数组
     * @param feeAmounts 手续费数组
     * @param userData 用户数据
     */
    function receiveFlashLoan(
        IERC20[] memory tokens,
        uint256[] memory amounts,
        uint256[] memory feeAmounts,
        bytes memory userData
    ) external;
}
```

---

## 4. Aave 集成

### 4.1 Aave V3 Pool 接口

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

interface IPool {
    /**
     * @dev 发起 Flash Loan
     * @param receiverAddress 回调合约地址
     * @param assets 借款资产地址数组
     * @param amounts 借款金额数组
     * @param interestRateModes 利率模式（0: None, 1: Stable, 2: Variable）
     * @param onBehalfOf 代理人地址
     * @param params 参数（编码）
     * @param referralCode 推荐码
     */
    function flashLoan(
        address receiverAddress,
        address[] calldata assets,
        uint256[] calldata amounts,
        uint256[] calldata interestRateModes,
        address onBehalfOf,
        bytes calldata params,
        uint16 referralCode
    ) external;

    /**
     * @dev 获取 Flash Loan 手续费
     * @param asset 资产地址
     * @return 手续费（1e27 = 100%）
     */
    function FLASH_LOAN_PREMIUM_TOTAL() external view returns (uint128);
}
```

### 4.2 Aave 合约地址

| 网络 | Pool 地址 |
|------|----------|
| Ethereum Mainnet | `0x8787B6dF9124118DF7435e05DF517C7cA74d3B30` |
| Goerli Testnet | `0x5041c7539368C6665DF49DA1c529b275D4A6b591` |
| Polygon | `0x794a61358D6845594F94dc1DB02A252b5b4814aD` |

### 4.3 Aave 集成实现

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {IPool} from "@aave/v3-core/contracts/interfaces/IPool.sol";
import {IFlashLoanReceiver} from "./interfaces/IFlashLoanReceiver.sol";

/**
 * @title AaveFlashLoan
 * @author yangyangyang
 * @notice Aave Flash Loan 实现
 */
contract AaveFlashLoan is IFlashLoanReceiver {
    // Aave Pool 地址（主网）
    IPool public constant POOL =
        IPool(0x8787B6dF9124118DF7435e05DF517C7cA74d3B30);

    // 所有者
    address public owner;

    // 事件
    event FlashLoanExecuted(
        address indexed asset,
        uint256 amount,
        uint256 premium
    );

    // 修饰符
    modifier onlyOwner() {
        require(msg.sender == owner, "Not owner");
        _;
    }

    /**
     * @dev 构造函数
     */
    constructor() {
        owner = msg.sender;
    }

    /**
     * @dev 发起 Flash Loan（对外接口）
     * @param asset 借款资产地址
     * @param amount 借款金额
     */
    function requestFlashLoan(
        address asset,
        uint256 amount
    ) external onlyOwner {
        address[] memory assets = new address[](1);
        assets[0] = asset;

        uint256[] memory amounts = new uint256[](1);
        amounts[0] = amount;

        uint256[] memory modes = new uint256[](1);
        modes[0] = 0;

        // 编码参数
        bytes memory params = abi.encode(msg.sender, asset, amount);

        // 调用 Aave Flash Loan
        POOL.flashLoan(
            address(this),
            assets,
            amounts,
            modes,
            address(this),
            params,
            0 // referralCode
        );
    }

    /**
     * @dev Aave Flash Loan 回调函数
     * @param asset 借款资产地址
     * @param amount 借款金额
     * @param premium 手续费
     * @param initiator 发起人地址
     * @param params 参数（编码）
     * @return 成功返回 true
     */
    function executeOperation(
        address asset,
        uint256 amount,
        uint256 premium,
        address initiator,
        bytes calldata params
    ) external override returns (bool) {
        // 验证调用者
        require(msg.sender == address(POOL), "Invalid caller");
        require(initiator == owner, "Invalid initiator");

        // 解码参数
        (
            address user,
            address tokenBorrow,
            uint256 amountBorrow
        ) = abi.decode(params, (address, address, uint256));

        // 1. 执行套利逻辑
        _doArbitrage(asset, amount, premium);

        // 2. 批准 Aave Pool 扣除本金 + 手续费
        uint256 amountOwed = amount + premium;
        IERC20(asset).approve(address(POOL), amountOwed);

        emit FlashLoanExecuted(asset, amount, premium);

        return true;
    }

    /**
     * @dev 执行套利逻辑
     * @param asset 借款资产
     * @param amount 借款金额
     * @param premium 手续费
     */
    function _doArbitrage(
        address asset,
        uint256 amount,
        uint256 premium
    ) internal {
        // TODO: 实现具体套利逻辑
        // 1. 在 DEX A 买入代币
        // 2. 在 DEX B 卖出代币
        // 3. 计算利润
        // 4. 确保利润 > 手续费

        // 示例：在 Uniswap 买入，在 SushiSwap 卖出
        // ...

        // 确保利润足够支付手续费
        uint256 profit = _calculateProfit(asset, amount);
        require(profit > premium, "Profit insufficient");
    }

    /**
     * @dev 计算利润
     * @param asset 资产地址
     * @param amount 金额
     * @return 利润
     */
    function _calculateProfit(
        address asset,
        uint256 amount
    ) internal pure returns (uint256 profit) {
        // TODO: 实现利润计算逻辑
        return amount * 5 / 1000; // 示例：0.5% 利润
    }

    /**
     * @dev 提取资金
     * @param asset 资产地址
     * @param amount 金额
     */
    function withdraw(address asset, uint256 amount) external onlyOwner {
        IERC20(asset).transfer(owner, amount);
    }

    /**
     * @dev 紧急暂停
     */
    function pause() external onlyOwner {
        // TODO: 实现 Pausable
    }

    /**
     * @dev 接收 ETH
     */
    receive() external payable {}
}
```

---

## 5. Uniswap V3 Flash

### 5.1 Uniswap V3 Pool 接口

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

interface IUniswapV3Pool {
    /**
     * @dev 发起 Flash Loan
     * @param recipient 接收地址
     * @param amount0 token0 借款金额
     * @param amount1 token1 借款金额
     * @param data 参数（编码）
     */
    function flash(
        address recipient,
        uint256 amount0,
        uint256 amount1,
        bytes calldata data
    ) external;
}
```

### 5.2 Uniswap V3 集成实现

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {IUniswapV3Pool} from "@uniswap/v3-core/contracts/interfaces/IUniswapV3Pool.sol";
import {IUniswapV3FlashCallback} from "./interfaces/IUniswapV3FlashCallback.sol";

/**
 * @title UniswapV3FlashLoan
 * @author yangyangyang
 * @notice Uniswap V3 Flash Loan 实现
 */
contract UniswapV3FlashLoan is IUniswapV3FlashCallback {
    // Uniswap V3 Pool 地址（USDC/ETH 0.3%）
    IUniswapV3Pool public constant POOL =
        IUniswapV3Pool(0x8ad599c3A0ff1De082011EFDDc58f1908eb6e6D8);

    // 所有者
    address public owner;

    // 事件
    event FlashLoanExecuted(
        uint256 amount0,
        uint256 amount1,
        uint256 fee0,
        uint256 fee1
    );

    /**
     * @dev 构造函数
     */
    constructor() {
        owner = msg.sender;
    }

    /**
     * @dev 发起 Flash Loan（对外接口）
     * @param amount0 token0 借款金额
     * @param amount1 token1 借款金额
     */
    function requestFlashLoan(
        uint256 amount0,
        uint256 amount1
    ) external {
        require(msg.sender == owner, "Not owner");

        // 编码参数
        bytes memory data = abi.encode(msg.sender, amount0, amount1);

        // 调用 Uniswap V3 Flash
        POOL.flash(address(this), amount0, amount1, data);
    }

    /**
     * @dev Uniswap V3 Flash Loan 回调函数
     * @param fee0 token0 手续费
     * @param fee1 token1 手续费
     * @param data 参数（编码）
     */
    function uniswapV3FlashCallback(
        uint256 fee0,
        uint256 fee1,
        bytes calldata data
    ) external override {
        // 验证调用者
        require(msg.sender == address(POOL), "Invalid caller");

        // 解码参数
        (address user, uint256 amount0, uint256 amount1) = abi.decode(
            data,
            (address, uint256, uint256)
        );

        // 1. 执行套利逻辑
        _doArbitrage(amount0, amount1, fee0, fee1);

        // 2. 归还本金 + 手续费
        if (amount0 > 0) {
            IERC20(POOL.token0()).transfer(
                msg.sender,
                amount0 + fee0
            );
        }
        if (amount1 > 0) {
            IERC20(POOL.token1()).transfer(
                msg.sender,
                amount1 + fee1
            );
        }

        emit FlashLoanExecuted(amount0, amount1, fee0, fee1);
    }

    /**
     * @dev 执行套利逻辑
     */
    function _doArbitrage(
        uint256 amount0,
        uint256 amount1,
        uint256 fee0,
        uint256 fee1
    ) internal {
        // TODO: 实现具体套利逻辑
    }

    /**
     * @dev 接收 ETH
     */
    receive() external payable {}
}
```

---

## 6. Balancer Flash

### 6.1 Balancer Vault 接口

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

interface IBalancerVault {
    /**
     * @dev 发起 Flash Loan
     * @param recipient 接收地址
     * @param tokens 借款代币数组
     * @param amounts 借款金额数组
     * @param userData 用户数据
     */
    function flashLoan(
        address recipient,
        address[] memory tokens,
        uint256[] memory amounts,
        bytes memory userData
    ) external;
}
```

### 6.2 Balancer 集成实现

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {IBalancerVault} from "@balancer-labs/v2-interfaces/contracts/vault/IVault.sol";
import {IVaultFlashLoanReceiver} from "./interfaces/IVaultFlashLoanReceiver.sol";

/**
 * @title BalancerFlashLoan
 * @author yangyangyang
 * @notice Balancer Flash Loan 实现
 */
contract BalancerFlashLoan is IVaultFlashLoanReceiver {
    // Balancer Vault 地址（主网）
    IBalancerVault public constant VAULT =
        IBalancerVault(0xBA12222222228d8Ba445958a75a0704d566BF2C8);

    // 所有者
    address public owner;

    /**
     * @dev 构造函数
     */
    constructor() {
        owner = msg.sender;
    }

    /**
     * @dev 发起 Flash Loan（对外接口）
     * @param tokens 借款代币数组
     * @param amounts 借款金额数组
     */
    function requestFlashLoan(
        address[] memory tokens,
        uint256[] memory amounts
    ) external {
        require(msg.sender == owner, "Not owner");

        // 编码参数
        bytes memory userData = abi.encode(msg.sender, tokens, amounts);

        // 调用 Balancer Flash
        VAULT.flashLoan(address(this), tokens, amounts, userData);
    }

    /**
     * @dev Balancer Flash Loan 回调函数
     * @param tokens 借款代币数组
     * @param amounts 借款金额数组
     * @param feeAmounts 手续费数组
     * @param userData 用户数据
     */
    function receiveFlashLoan(
        IERC20[] memory tokens,
        uint256[] memory amounts,
        uint256[] memory feeAmounts,
        bytes memory userData
    ) external override {
        // 验证调用者
        require(msg.sender == address(VAULT), "Invalid caller");

        // 解码参数
        (address user, , ) = abi.decode(userData, (address, address[], uint256[]));

        // 1. 执行套利逻辑
        _doArbitrage(tokens, amounts, feeAmounts);

        // 2. 归还本金 + 手续费
        for (uint256 i = 0; i < tokens.length; i++) {
            uint256 amountOwed = amounts[i] + feeAmounts[i];
            tokens[i].transfer(msg.sender, amountOwed);
        }
    }

    /**
     * @dev 执行套利逻辑
     */
    function _doArbitrage(
        IERC20[] memory tokens,
        uint256[] memory amounts,
        uint256[] memory feeAmounts
    ) internal {
        // TODO: 实现具体套利逻辑
    }
}
```

---

## 7. Solidity 实现

### 7.1 完整套利合约

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {ReentrancyGuard} from "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {IPool} from "@aave/v3-core/contracts/interfaces/IPool.sol";

/**
 * @title FlashLoanArbitrage
 * @author yangyangyang
 * @notice Flash Loan 套利合约（集成 Aave）
 */
contract FlashLoanArbitrage is
    ReentrancyGuard,
    Ownable,
    IFlashLoanReceiver
{
    using SafeERC20 for IERC20;

    // Aave Pool
    IPool public constant AAVE_POOL =
        IPool(0x8787B6dF9124118DF7435e05DF517C7cA74d3B30);

    // 套利参数
    struct ArbitrageParams {
        address tokenIn;      // 输入代币
        address tokenOut;     // 输出代币
        uint256 amountIn;     // 输入金额
        address dexA;         // DEX A 地址
        address dexB;         // DEX B 地址
        uint256 minProfit;    // 最小利润
    }

    // 统计数据
    uint256 public totalArbitrages;     // 总套利次数
    uint256 public totalProfit;         // 总利润
    uint256 public totalFeesPaid;       // 总手续费

    // 事件
    event ArbitrageExecuted(
        address indexed tokenIn,
        address indexed tokenOut,
        uint256 amountIn,
        uint256 profit,
        uint256 fees
    );

    /**
     * @dev 构造函数
     */
    constructor() Ownable(msg.sender) {}

    /**
     * @dev 发起 Flash Loan 套利
     * @param params 套利参数
     */
    function executeArbitrage(ArbitrageParams calldata params)
        external
        onlyOwner
        nonReentrant
    {
        // 验证参数
        require(params.tokenIn != address(0), "Invalid tokenIn");
        require(params.tokenOut != address(0), "Invalid tokenOut");
        require(params.amountIn > 0, "Invalid amountIn");

        // 准备借款参数
        address[] memory assets = new address[](1);
        assets[0] = params.tokenIn;

        uint256[] memory amounts = new uint256[](1);
        amounts[0] = params.amountIn;

        uint256[] memory modes = new uint256[](1);
        modes[0] = 0;

        // 编码参数
        bytes memory userData = abi.encode(params);

        // 发起 Flash Loan
        AAVE_POOL.flashLoan(
            address(this),
            assets,
            amounts,
            modes,
            address(this),
            userData,
            0
        );
    }

    /**
     * @dev Aave Flash Loan 回调
     */
    function executeOperation(
        address asset,
        uint256 amount,
        uint256 premium,
        address initiator,
        bytes calldata params
    ) external override returns (bool) {
        // 验证调用者
        require(msg.sender == address(AAVE_POOL), "Invalid caller");
        require(initiator == owner(), "Invalid initiator");

        // 解码参数
        ArbitrageParams memory arbitrageParams = abi.decode(
            params,
            (ArbitrageParams)
        );

        // 1. 执行套利
        uint256 profit = _executeArbitrageLogic(
            arbitrageParams,
            amount,
            premium
        );

        // 2. 归还借款
        uint256 amountOwed = amount + premium;
        IERC20(asset).safeApprove(address(AAVE_POOL), amountOwed);

        // 3. 更新统计
        totalArbitrages++;
        totalProfit += profit;
        totalFeesPaid += premium;

        emit ArbitrageExecuted(
            arbitrageParams.tokenIn,
            arbitrageParams.tokenOut,
            amount,
            profit,
            premium
        );

        return true;
    }

    /**
     * @dev 执行套利逻辑
     */
    function _executeArbitrageLogic(
        ArbitrageParams memory params,
        uint256 amount,
        uint256 premium
    ) internal returns (uint256 profit) {
        // 1. 在 DEX A 买入 tokenOut
        uint256 amountOut = _swapOnDEX(
            params.dexA,
            params.tokenIn,
            params.tokenOut,
            amount
        );

        // 2. 在 DEX B 卖出 tokenOut
        uint256 amountBack = _swapOnDEX(
            params.dexB,
            params.tokenOut,
            params.tokenIn,
            amountOut
        );

        // 3. 计算利润
        profit = amountBack - amount - premium;

        // 验证利润
        require(profit >= params.minProfit, "Profit too low");

        return profit;
    }

    /**
     * @dev 在 DEX 上执行 Swap
     */
    function _swapOnDEX(
        address dex,
        address tokenIn,
        address tokenOut,
        uint256 amountIn
    ) internal returns (uint256 amountOut) {
        // TODO: 实现具体 DEX Swap 逻辑
        // 示例：调用 Uniswap V2 Router
        // IUniswapV2Router(dex).swapExactTokensForTokens(
        //     amountIn,
        //     0,
        //     getPath(tokenIn, tokenOut),
        //     address(this),
        //     block.timestamp
        // );

        return amountIn * 1005 / 1000; // 示例：0.5% 收益
    }

    /**
     * @dev 提取利润
     */
    function withdrawProfit(address token, uint256 amount) external onlyOwner {
        IERC20(token).safeTransfer(owner(), amount);
    }

    /**
     * @dev 紧急提取
     */
    function emergencyWithdraw(address token) external onlyOwner {
        uint256 balance = IERC20(token).balanceOf(address(this));
        IERC20(token).safeTransfer(owner(), balance);
    }
}
```

---

## 8. 安全考虑

### 8.1 重入攻击防护

```solidity
import {ReentrancyGuard} from "@openzeppelin/contracts/security/ReentrancyGuard.sol";

contract FlashLoanArbitrage is ReentrancyGuard {
    function executeOperation(...) external nonReentrant {
        // 逻辑
    }
}
```

### 8.2 访问控制

```solidity
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";

contract FlashLoanArbitrage is Ownable, AccessControl {
    bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");

    function executeArbitrage(...) external onlyRole(OPERATOR_ROLE) {
        // 逻辑
    }
}
```

### 8.3 输入验证

```solidity
function executeArbitrage(ArbitrageParams calldata params) external {
    // 验证地址
    require(params.tokenIn != address(0), "Invalid tokenIn");
    require(params.tokenOut != address(0), "Invalid tokenOut");
    require(params.dexA != address(0), "Invalid dexA");
    require(params.dexB != address(0), "Invalid dexB");

    // 验证金额
    require(params.amountIn > 0, "Invalid amountIn");
    require(params.minProfit > 0, "Invalid minProfit");

    // 验证代币合约
    require(
        IERC20(params.tokenIn).totalSupply() > 0,
        "Invalid tokenIn contract"
    );
}
```

### 8.4 滑点保护

```solidity
function _swapWithSlippageProtection(
    uint256 amountIn,
    uint256 minAmountOut
) internal returns (uint256 amountOut) {
    uint256 amountOut = _doSwap(amountIn);

    require(
        amountOut >= minAmountOut,
        "Slippage exceeded"
    );

    return amountOut;
}
```

---

## 9. 部署和测试

### 9.1 Hardhat 配置

```javascript
// hardhat.config.js
require("@nomicfoundation/hardhat-toolbox");
require("@nomicfoundation/hardhat-verify");
require("dotenv").config();

module.exports = {
  solidity: {
    version: "0.8.20",
    settings: {
      optimizer: {
        enabled: true,
        runs: 200,
      },
    },
  },
  networks: {
    hardhat: {
      forking: {
        url: process.env.MAINNET_RPC_URL,
      },
    },
    goerli: {
      url: process.env.GOERLI_RPC_URL,
      accounts: [process.env.PRIVATE_KEY],
    },
    mainnet: {
      url: process.env.MAINNET_RPC_URL,
      accounts: [process.env.PRIVATE_KEY],
    },
  },
  etherscan: {
    apiKey: process.env.ETHERSCAN_API_KEY,
  },
};
```

### 9.2 部署脚本

```javascript
// scripts/deploy.js
const hre = require("hardhat");

async function main() {
  const FlashLoanArbitrage = await hre.ethers.getContractFactory(
    "FlashLoanArbitrage"
  );
  const contract = await FlashLoanArbitrage.deploy();

  await contract.deployed();

  console.log("FlashLoanArbitrage deployed to:", contract.address);

  // 验证合约（Etherscan）
  if (hre.network.name !== "hardhat" && hre.network.name !== "localhost") {
    await hre.run("verify:verify", {
      address: contract.address,
      constructorArguments: [],
    });
  }
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
```

### 9.3 测试用例

```javascript
// test/FlashLoanArbitrage.test.js
const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("FlashLoanArbitrage", function () {
  let contract;
  let owner;
  let addr1;

  beforeEach(async function () {
    [owner, addr1] = await ethers.getSigners();

    const FlashLoanArbitrage = await ethers.getContractFactory(
      "FlashLoanArbitrage"
    );
    contract = await FlashLoanArbitrage.deploy();
    await contract.deployed();
  });

  it("Should execute arbitrage successfully", async function () {
    const params = {
      tokenIn: "0x...", // USDC address
      tokenOut: "0x...", // USDT address
      amountIn: ethers.utils.parseUnits("10000", 6), // 10000 USDC
      dexA: "0x...", // Uniswap V2 Router
      dexB: "0x...", // SushiSwap Router
      minProfit: ethers.utils.parseUnits("10", 6), // 10 USDC
    };

    await expect(contract.executeArbitrage(params))
      .to.emit(contract, "ArbitrageExecuted");
  });

  it("Should fail if profit is too low", async function () {
    const params = {
      tokenIn: "0x...",
      tokenOut: "0x...",
      amountIn: ethers.utils.parseUnits("10000", 6),
      dexA: "0x...",
      dexB: "0x...",
      minProfit: ethers.utils.parseUnits("1000", 6), // 太高的最小利润
    };

    await expect(
      contract.executeArbitrage(params)
    ).to.be.revertedWith("Profit too low");
  });

  it("Should prevent unauthorized access", async function () {
    const params = { /* ... */ };

    await expect(
      contract.connect(addr1).executeArbitrage(params)
    ).to.be.revertedWith("Not owner");
  });
});
```

---

## 10. Gas 优化

### 10.1 优化技巧

**1. 使用 `calldata` 代替 `memory`**

```solidity
// ✓ 优化前
function execute(ArbitrageParams memory params) external {
    // Gas: ~50000
}

// ✓ 优化后
function execute(ArbitrageParams calldata params) external {
    // Gas: ~47000（节省 ~3000 gas）
}
```

**2. 批量操作**

```solidity
// ✓ 批量授权
IERC20(token).safeApprove(dexA, amount);
IERC20(token).safeApprove(dexB, amount);

// 优化为批量操作
IERC20(token).safeApprove(dexA, type(uint256).max);
```

**3. 使用 `unchecked`（Solidity 0.8+）**

```solidity
// ✓ 优化前
uint256 total = amount0 + amount1;

// ✓ 优化后（如果确定不会溢出）
uint256 total;
unchecked {
    total = amount0 + amount1;
}
```

### 10.2 Gas 预估

```solidity
// 估算 Gas
function estimateGas(ArbitrageParams calldata params)
    external
    view
    returns (uint256)
{
    // Flash Loan 基础 Gas: ~200000
    // Swap 操作 Gas: ~150000
    // 合约调用 Gas: ~50000
    // 总计: ~400000 gas

    return 400000;
}
```

---

## 附录

### A. 相关文档

- [Blockchain_TechStack.md](../TechStack/Blockchain_TechStack.md) - 区块链技术栈
- [DEX_Monitor.md](./DEX_Monitor.md) - DEX 监控模块
- [MEV_Engine.md](./MEV_Engine.md) - MEV 引擎

### B. 外部资源

- [Aave 开发者文档](https://docs.aave.com/developers)
- [Uniswap V3 文档](https://docs.uniswap.org/protocol/V3/introduction)
- [Balancer 文档](https://docs.balancer.fi/)

### C. 常见问题

**Q1: Flash Loan 会失败吗？**
A: 会。如果业务逻辑执行失败或资金未归还，整个交易会回滚。

**Q2: 手续费如何计算？**
A: Aave: 0.05%, Uniswap V3: 0.3%, Balancer: 0%（可能包含交易滑点）

**Q3: 如何测试 Flash Loan？**
A: 使用 Hardhat fork 主网进行测试，或使用 Aave 测试网。

**Q4: Flash Loan 安全吗？**
A: 智能合约本身安全，但需注意业务逻辑安全（重入、访问控制等）。

---

**最后更新**: 2026-01-07
**版本**: v1.0.0
