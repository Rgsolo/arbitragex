# Task 持久化和恢复机制

**版本**: v1.0.0
**创建日期**: 2026-01-08

---

## 1. Task 的工作原理

### 1.1 Task Tool 说明

`Task()` 是 Claude Code 的内置工具，用于启动子代理处理任务。

**位置**：内置在 Claude Code 中，无需额外安装

**参数**：
```python
Task(
    subagent_type: str,        # agent 类型
    prompt: str,               # 任务描述
    model: str = "sonnet",     # 模型选择 (sonnet/opus/haiku)
    description: str,          # 简短描述（3-5个词）
    run_in_background: bool = False,  # 是否后台运行
    resume: str = None         # 恢复之前的 task ID
)
```

### 1.2 运行模式

**前台运行（默认）**：
```python
# 阻塞等待任务完成
result = Task(
    subagent_type="general-purpose",
    prompt="创建项目结构",
    model="sonnet"
)
# 任务完成后继续，result 包含结果
```

**后台运行（并行）**：
```python
# 立即返回 task_id，任务在后台运行
task_id = Task(
    subagent_type="general-purpose",
    prompt="创建项目结构",
    model="sonnet",
    run_in_background=True
)
# task_id: "task-abc123"
```

**获取后台任务结果**：
```python
# 使用 TaskOutput 获取结果
result = TaskOutput(
    task_id="task-abc123",
    block=True,  # 阻塞等待完成
    timeout=300000  # 超时时间（毫秒）
)
```

---

## 2. 持久化问题

### 2.1 问题说明

❌ **Task 本身不持久化**：
- Task 运行状态存储在内存中
- 电脑关机 → 后台 Task **丢失**
- 无法自动恢复中断的 Task

### 2.2 解决方案

✅ **使用文件持久化进度**：
- 创建 `.parallel-tasks.json` 跟踪任务状态
- 每 5 分钟自动保存一次进度
- 关机后可从文件恢复

---

## 3. 持久化方案设计

### 3.1 进度文件结构

`.parallel-tasks.json`：

```json
{
  "session_id": "phase2-session-20260108",
  "phase": "Phase 2: 基础架构搭建",
  "status": "in_progress",
  "started_at": "2026-01-08 10:00:00",
  "last_updated": "2026-01-08 10:05:00",

  "parallel_tasks": [
    {
      "task_id": "2.1",
      "name": "项目结构初始化",
      "agent_role": "go-developer",
      "status": "completed",
      "started_at": "2026-01-08 10:00:00",
      "completed_at": "2026-01-08 10:02:00",
      "claude_task_id": "task-abc123",
      "result": {
        "files_created": [
          "cmd/price/main.go",
          "internal/config/config.go"
        ],
        "notes": "项目结构已按 go-zero 规范初始化"
      }
    },
    {
      "task_id": "2.2",
      "name": "Docker 环境配置",
      "agent_role": "devops-engineer",
      "status": "in_progress",
      "started_at": "2026-01-08 10:00:00",
      "claude_task_id": "task-def456"
    },
    {
      "task_id": "2.3",
      "name": "MySQL 数据库部署",
      "agent_role": "devops-engineer",
      "status": "pending"
    }
  ],

  "recovery_info": {
    "last_checkpoint": "任务 2.1 已完成，2.2 进行中，2.3 待开始",
    "next_action": "等待任务 2.2 完成，然后启动任务 2.3",
    "can_resume": true
  }
}
```

### 3.2 状态定义

| 状态 | 说明 | 可恢复 |
|------|------|--------|
| `pending` | 未开始 | ✅ 从头开始 |
| `in_progress` | 进行中 | ⚠️ 需重新启动（会丢失进度） |
| `completed` | 已完成 | ✅ 跳过，使用已有结果 |
| `failed` | 失败 | ✅ 重新启动 |

---

## 4. 工作流程设计

### 4.1 标准并行开发流程

```python
# 步骤 1: 读取/创建进度文件
progress = load_progress()

# 步骤 2: 启动并行任务
task_ids = []
for task in progress.parallel_tasks:
    if task.status == "pending":
        # 启动新任务
        task_id = Task(
            subagent_type="general-purpose",
            prompt=task.prompt,
            run_in_background=True
        )
        task_ids.append(task_id)

        # 更新进度
        task.status = "in_progress"
        task.started_at = now()
        task.claude_task_id = task_id
        save_progress(progress)

# 步骤 3: 等待所有任务完成
for task_id in task_ids:
    result = TaskOutput(task_id, block=True)

    # 更新进度
    update_task_progress(task_id, result)
    save_progress(progress)
```

### 4.2 恢复流程

**关机后恢复**：

```python
# 步骤 1: 读取进度文件
progress = load_progress(".parallel-tasks.json")

# 步骤 2: 检查状态
print(f"Session: {progress.session_id}")
print(f"Last updated: {progress.last_updated}")
print(f"Recovery info: {progress.recovery_info.last_checkpoint}")

# 步骤 3: 恢复未完成的任务
for task in progress.parallel_tasks:
    if task.status == "in_progress":
        # 重新启动（会丢失之前的进度）
        print(f"Resuming task: {task.task_id} - {task.name}")
        task_id = Task(
            subagent_type=task.agent_type,
            prompt=task.prompt,
            run_in_background=True
        )
        task.claude_task_id = task_id

    elif task.status == "pending":
        # 启动新任务
        print(f"Starting task: {task.task_id} - {task.name}")
        task_id = Task(...)
        task.claude_task_id = task_id

    elif task.status == "completed":
        # 跳过已完成任务
        print(f"Skipping completed task: {task.task_id}")

# 步骤 4: 保存新的进度
save_progress(progress)
```

---

## 5. 实际使用示例

### 5.1 启动 Phase 2 并行开发

```python
# === 主 Agent 执行 ===

# 1. 读取进度文件
progress = load_or_create_progress(".parallel-tasks.json")

# 2. 定义任务
tasks = [
    {
        "task_id": "2.1",
        "name": "项目结构初始化",
        "agent_role": "go-developer",
        "prompt": """
        你是 Go 开发工程师，使用 go-zero 框架初始化项目结构。

        要求：
        - 创建 cmd/price/main.go
        - 创建 internal/config/config.go
        - 初始化 go.mod 和 go.sum
        - 遵循 go-zero 项目结构规范

        参考：
        - docs/design/Architecture/Module_Structure.md
        - CLAUDE.md 中的 go-zero 最佳实践
        """,
        "deliverables": [
            "cmd/price/main.go",
            "internal/config/config.go",
            "go.mod"
        ]
    },
    {
        "task_id": "2.2",
        "name": "Docker 环境配置",
        "agent_role": "devops-engineer",
        "prompt": """
        你是 DevOps 工程师，配置 Docker 环境。

        要求：
        - 创建 docker-compose.yml（包含 MySQL, Redis, 3 个服务）
        - 创建 3 个 Dockerfile（price, engine, trade）
        - 配置健康检查和资源限制

        参考：
        - docs/design/Deployment/Docker_Deployment.md
        - CLAUDE.md 中的 Docker 配置
        """,
        "deliverables": [
            "docker-compose.yml",
            "Dockerfile.price",
            "Dockerfile.engine",
            "Dockerfile.trade"
        ]
    },
    {
        "task_id": "2.3",
        "name": "MySQL 数据库部署",
        "agent_role": "devops-engineer",
        "prompt": """
        你是 DevOps 工程师，部署 MySQL 数据库。

        要求：
        - 创建初始化脚本 scripts/mysql/01-init-database.sql
        - 创建配置文件 config/mysql.cnf
        - 包含所有必需的表结构

        参考：
        - docs/design/Database/Schema_Design.md
        """,
        "deliverables": [
            "scripts/mysql/01-init-database.sql",
            "config/mysql.cnf"
        ]
    }
]

# 3. 保存任务定义到进度文件
progress.parallel_tasks = tasks
save_progress(progress)

# 4. 并行启动 3 个任务
task_ids = []
for task in tasks:
    print(f"启动任务: {task['name']}")

    task_id = Task(
        subagent_type="general-purpose",
        prompt=task['prompt'],
        description=task['name'],
        model="sonnet",
        run_in_background=True
    )

    task_ids.append(task_id)

    # 更新进度
    task['status'] = 'in_progress'
    task['started_at'] = now()
    task['claude_task_id'] = task_id

    save_progress(progress)  # 每启动一个任务就保存一次

# 5. 等待所有任务完成
print("等待所有任务完成...")
results = []
for task_id in task_ids:
    result = TaskOutput(task_id, block=True, timeout=300000)
    results.append(result)

    # 更新进度
    task = find_task_by_claude_id(progress, task_id)
    task['status'] = 'completed'
    task['completed_at'] = now()
    task['result'] = result

    save_progress(progress)  # 每完成一个任务就保存一次

# 6. 显示结果
for i, result in enumerate(results):
    print(f"任务 {tasks[i]['name']} 完成:")
    print(f"  结果: {result}")
```

### 5.2 关机后恢复

**场景**：电脑关机了，重新打开 Claude Code

```python
# === 主 Agent 执行 ===

# 1. 读取进度文件
progress = load_progress(".parallel-tasks.json")

# 2. 显示恢复信息
print("=== 恢复 Session ===")
print(f"Session: {progress.session_id}")
print(f"Phase: {progress.phase}")
print(f"最后更新: {progress.last_updated}")
print(f"恢复信息: {progress.recovery_info.last_checkpoint}")
print()

# 3. 检查任务状态
print("=== 任务状态 ===")
for task in progress.parallel_tasks:
    status_emoji = {
        "pending": "⏳",
        "in_progress": "🔄",
        "completed": "✅",
        "failed": "❌"
    }[task.status]

    print(f"{status_emoji} [{task['task_id']}] {task['name']}: {task.status}")

    if task.status == "completed":
        print(f"   产出: {task['result'].files_created}")
    elif task.status == "in_progress":
        print(f"   ⚠️  任务进行中，需要重新启动")

# 4. 询问用户
print()
print("发现未完成的任务，是否继续？")

# 5. 恢复未完成的任务
incomplete_tasks = [
    t for t in progress.parallel_tasks
    if t.status in ["pending", "in_progress"]
]

if incomplete_tasks:
    print(f"发现 {len(incomplete_tasks)} 个未完成任务，继续执行...")

    task_ids = []
    for task in incomplete_tasks:
        print(f"重新启动任务: {task['name']}")

        task_id = Task(
            subagent_type="general-purpose",
            prompt=task['prompt'],
            description=task['name'],
            model="sonnet",
            run_in_background=True
        )

        task_ids.append(task_id)
        task['status'] = 'in_progress'
        task['started_at'] = now()
        task['claude_task_id'] = task_id

        save_progress(progress)

    # 等待完成...
    for task_id in task_ids:
        result = TaskOutput(task_id, block=True)
        update_and_save(progress, task_id, result)

else:
    print("✅ 所有任务已完成！")
```

---

## 6. 最佳实践

### 6.1 进度保存策略

✅ **频繁保存**：
- 每启动一个任务 → 保存一次
- 每完成一个任务 → 保存一次
- 每隔 5 分钟 → 自动保存一次

✅ **原子更新**：
```python
# 先写到临时文件，再重命名（保证原子性）
tmp_file = ".parallel-tasks.json.tmp"
save_to_file(tmp_file, progress)
os.rename(tmp_file, ".parallel-tasks.json")
```

### 6.2 关机前准备

✅ **检查点**：
```python
# 关机前，记录当前状态
progress.recovery_info.last_checkpoint = "准备关机，所有任务已保存"
progress.recovery_info.can_resume = True
save_progress(progress)
print("✅ 进度已保存，可随时恢复")
```

### 6.3 恢复后检查

✅ **验证文件**：
```python
# 恢复后，检查已创建的文件是否存在
for task in progress.parallel_tasks:
    if task.status == "completed":
        for file_path in task.result.files_created:
            if not os.path.exists(file_path):
                print(f"⚠️  文件不存在: {file_path}")
                task.status = "failed"  # 标记为失败，重新执行
```

---

## 7. 工具函数

### 7.1 进度管理

```python
import json
from datetime import datetime
from pathlib import Path

def load_or_create_progress(file_path):
    """加载或创建进度文件"""
    if Path(file_path).exists():
        with open(file_path, 'r') as f:
            return json.load(f)
    else:
        return {
            "session_id": f"session-{datetime.now().strftime('%Y%m%d%H%M%S')}",
            "phase": "Unknown",
            "status": "in_progress",
            "started_at": datetime.now().isoformat(),
            "parallel_tasks": [],
            "recovery_info": {
                "can_resume": True
            }
        }

def save_progress(progress, file_path=".parallel-tasks.json"):
    """保存进度到文件"""
    progress['last_updated'] = datetime.now().isoformat()
    with open(file_path, 'w') as f:
        json.dump(progress, f, indent=2)

def update_task_status(progress, claude_task_id, status, result=None):
    """更新任务状态"""
    for task in progress['parallel_tasks']:
        if task.get('claude_task_id') == claude_task_id:
            task['status'] = status
            if status == 'completed':
                task['completed_at'] = datetime.now().isoformat()
                task['result'] = result
            save_progress(progress)
            break
```

---

## 8. 总结

### 8.1 关键要点

1. **Task 不持久化**：电脑关机会丢失后台任务
2. **解决方案**：使用 `.parallel-tasks.json` 跟踪进度
3. **频繁保存**：每启动/完成一个任务就保存一次
4. **可恢复**：关机后可从进度文件恢复未完成任务

### 8.2 工作流程

```
启动任务
  ↓
保存到 .parallel-tasks.json (status=in_progress)
  ↓
关机（任务丢失，但进度文件保留）
  ↓
重启，读取 .parallel-tasks.json
  ↓
重新启动未完成的任务
  ↓
更新进度文件 (status=completed)
```

### 8.3 文件清单

- `.parallel-tasks.json` - 并行任务进度跟踪
- `.progress.json` - 项目整体进度跟踪
- `docs/development/TASK_RECOVERY.md` - 本文档

---

**文档版本**: v1.0.0
**最后更新**: 2026-01-08
