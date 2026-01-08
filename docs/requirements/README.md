# ArbitrageX 产品需求文档

本文档已模块化拆分，便于维护和更新。

## 📚 文档结构

### 核心文档（必读）
- **[PRD_Core.md](./PRD_Core.md)** - 核心产品需求
- **[PRD_Technical.md](./PRD_Technical.md)** - 技术需求
- **[PRD_Implementation.md](./PRD_Implementation.md)** - 实施计划

### 策略文档
详见 [Strategies/README.md](./Strategies/README.md)

### 历史版本
详见 [Archives/README.md](./Archives/README.md)

---

## 🔍 快速查找

### 我想了解...

| 主题 | 文档 |
|------|------|
| **产品全貌** | [PRD_Core.md](./PRD_Core.md) |
| **技术约束** | [PRD_Technical.md](./PRD_Technical.md) |
| **开发计划** | [PRD_Implementation.md](./PRD_Implementation.md) |
| **CEX套利策略** | [Strategies/Strategy_CEX_Arbitrage.md](./Strategies/Strategy_CEX_Arbitrage.md) |
| **DEX套利策略** | [Strategies/Strategy_DEX_Arbitrage.md](./Strategies/Strategy_DEX_Arbitrage.md) |
| **Flash Loan实现** | [Strategies/Strategy_FlashLoan.md](./Strategies/Strategy_FlashLoan.md) |
| **MEV套利** | [Strategies/Strategy_MEV.md](./Strategies/Strategy_MEV.md) |

---

## 📊 版本对应关系

| 版本 | PRD_Core | Technical | Implementation | CEX策略 | DEX策略 | FlashLoan | MEV |
|------|----------|-----------|----------------|---------|---------|-----------|-----|
| v2.0 | v2.0.0   | v2.0.0    | v2.0.0         | v1.0.0  | v1.0.0  | v1.0.0     | v1.0.0 |

**版本管理策略**：
- **主版本号**（v2.0）：由PRD_Core维护，重大变更时统一升级
- **子版本号**（v2.0.1）：各文档独立维护，内容更新时独立升级

---

## 📝 更新日志

### v2.0 (2026-01-07)
- **重大重构**：拆分模块化文档结构
  - 将1782行的PRD拆分为8个模块化文档
  - 将1299行的DEX补充文档拆分到策略文档
- **优先级调整**：DEX套利提升至最高优先级 ⭐⭐⭐⭐⭐
- **新增文档**：
  - Flash Loan专项文档（Strategy_FlashLoan.md）
  - MEV套利专项文档（Strategy_MEV.md）
- **资金配置优化**：从10万USDT降至5万USDT（Flash Loan无需预持资金）
- **分阶段实施**：重新设计4个开发阶段

### v1.0 (2026-01-06)
- 初始版本
- 完成PRD和DEX补充文档

---

## 🚀 新手入门

1. **第一步**：阅读 [PRD_Core.md](./PRD_Core.md) 了解产品全貌
2. **第二步**：根据需要选择策略文档
   - 想了解CEX套利 → [Strategy_CEX_Arbitrage.md](./Strategies/Strategy_CEX_Arbitrage.md)
   - 想了解DEX套利 → [Strategy_DEX_Arbitrage.md](./Strategies/Strategy_DEX_Arbitrage.md)
3. **第三步**：开发前必读 [PRD_Technical.md](./PRD_Technical.md)
4. **第四步**：查看 [PRD_Implementation.md](./PRD_Implementation.md) 了解实施计划

---

## 🛠️ 维护指南

### 如何更新文档

**场景1：更新策略**
- 找到对应的策略文档（如Strategy_CEX_Arbitrage.md）
- 直接修改内容
- 更新版本号和变更日志
- 提交PR：`docs(strategy): 更新CEX套利阈值`

**场景2：更新技术需求**
- 修改PRD_Technical.md
- 检查是否影响策略文档
- 更新版本号
- 提交PR：`docs(technical): 添加PostgreSQL支持`

**场景3：调整里程碑**
- 更新PRD_Core.md的里程碑章节
- 调整PRD_Implementation.md的时间计划
- 更新版本号
- 提交PR：`docs(core): 调整里程碑计划`

### 文档更新清单

每次提交前检查：
- [ ] 文档长度在合理范围内（400-600行）
- [ ] 变更日志已更新
- [ ] 版本号已更新
- [ ] 交叉引用已验证
- [ ] Markdown格式正确
- [ ] 没有引入重复内容

---

## 📈 预期收益

### 短期（1个月内）
- **查找效率提升50%**：清晰的目录和导航
- **更新冲突减少80%**：职责边界清晰
- **新人上手时间缩短60%**：从核心到细节的渐进式阅读

### 中期（3-6个月）
- **策略迭代加速**：独立更新策略文档
- **文档质量提升**：聚焦主题，内容更深入
- **团队协作改善**：减少沟通成本

### 长期（6个月以上）
- **知识沉淀**：稳定的文档结构便于知识积累
- **扩展性**：新增策略/功能易于扩展
- **维护成本降低**：模块化结构降低维护复杂度

---

## 📧 联系方式

如有问题或建议，请：
1. 查阅相关文档
2. 在团队会议上讨论
3. 提交Issue或PR

---

**文档版本**: v2.0.0
**创建日期**: 2026-01-07
**最后更新**: 2026-01-07
**维护人**: ArbitrageX开发团队
