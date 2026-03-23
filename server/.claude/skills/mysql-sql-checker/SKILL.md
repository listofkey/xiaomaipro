---
name: mysql-sql-checker
description: >
  检查代码中的SQL语句和GORM用法是否符合MySQL规范。当用户要求检查SQL、审查数据库代码、
  或者说"检查sql"、"sql审查"、"检查数据库操作"时使用这个skill。也适用于用户说
  "review我的sql"、"看看数据库代码有没有问题"、"帮我检查下这段数据库操作"等场景。
---

# MySQL SQL Checker

检查Go代码中的SQL语句和GORM/GORM Gen用法，找出无法在MySQL上正确执行的问题。

## 检查流程

1. 确定检查范围 — 用户可能指定了具体文件，也可能想检查整个项目
2. 扫描代码，定位所有数据库操作
3. 逐项检查，标注问题
4. 输出结构化报告

## 检查范围

扫描以下两类数据库操作代码：

### 类型一：原生SQL字符串

在Go代码中搜索以下模式，提取其中的SQL语句：
- `db.Raw("...")` / `tx.Raw("...")`
- `db.Exec("...")` / `tx.Exec("...")`
- `db.Where("...")` 中包含完整SQL片段的情况
- 字符串变量赋值后传入上述方法的情况
- `.sql` 文件中的SQL语句

对这些SQL语句进行语法检查。

### 类型二：GORM 链式调用

扫描所有GORM和GORM Gen的链式调用，检查其是否能生成有效的MySQL语句。

## 检查规则

### 原生SQL检查项

1. **语法正确性**
   - SELECT/INSERT/UPDATE/DELETE 基本语法是否完整
   - 子查询、JOIN、UNION 结构是否正确
   - GROUP BY / HAVING / ORDER BY 使用是否合理

2. **MySQL 方言兼容性**
   - 是否使用了MySQL不支持的语法（如 PostgreSQL 的 `RETURNING`、SQLite 的 `UPSERT` 写法）
   - MySQL 保留字是否被正确反引号转义（如 `order`、`group`、`key` 作为列名）
   - 数据类型是否为MySQL支持的类型

3. **占位符和参数绑定**
   - `?` 占位符数量是否与传入参数数量匹配
   - 是否存在字符串拼接构造SQL的情况（SQL注入风险）

4. **表名和列名**
   - 如果项目有model定义（如GORM model），检查SQL中的表名/列名是否与model一致
   - 注意GORM默认的snake_case命名转换规则

### GORM 链式调用检查项

1. **方法使用正确性**
   - `.First()` / `.Find()` / `.Take()` 的使用场景是否正确
   - `.Save()` vs `.Create()` vs `.Updates()` 的语义区别是否被正确使用
   - `.Delete()` 搭配软删除（`gorm.DeletedAt`）时的行为是否符合预期

2. **Where 条件**
   - GORM Gen 的类型安全查询（如 `q.User.Phone.Eq(value)`）— 字段名是否存在于对应model
   - 原生 GORM 的字符串条件（如 `Where("user_id = ?", id)`）— 列名是否正确
   - 多条件组合是否正确（`.Where().Where()` 是 AND，`.Where().Or()` 是 OR）

3. **Update 操作**
   - `.Updates(map)` 中的 key 是否是有效的数据库列名（注意GORM用的是列名不是Go字段名）
   - `.Update("column", value)` 中 column 是否是有效列名
   - 是否遗漏了 `.Where()` 条件导致全表更新的风险

4. **事务使用**
   - `db.Transaction(func(tx *gorm.DB) error {...})` 内部操作是否都使用了 `tx` 而非外部的 `db`
   - 事务内的错误是否被正确返回（return err 而不是被忽略）

5. **Model 与 Table**
   - `.Table("name")` 中的表名是否与实际数据库表匹配
   - `.Model(&struct{})` 使用匿名struct时是否搭配了 `.Table()`
   - GORM Gen 的 `q.TableName` 是否引用了存在的表

6. **性能相关**
   - 是否缺少 `.WithContext(ctx)` 调用（影响超时控制和链路追踪）
   - `.Find()` 是否缺少分页或LIMIT限制（可能导致全表扫描）
   - N+1 查询问题 — 循环中重复查询而不是批量查询

## 输出格式

用以下格式输出检查结果：

### 无问题时
```
✅ SQL 检查通过，共检查了 N 处数据库操作，未发现问题。
```

### 发现问题时

对每个问题，给出：

```
❌ [严重程度] 文件路径:行号
   问题：简要描述
   代码：相关代码片段
   原因：为什么这是一个问题
   建议：如何修复
```

严重程度分三级：
- **🔴 错误** — SQL无法执行，或会导致数据错误（如全表更新、语法错误）
- **🟡 警告** — 代码能运行但存在隐患（如缺少WHERE条件、性能问题）
- **🟢 建议** — 可以改进但不影响正确性（如缺少WithContext）

最后给出一个汇总：
```
📊 检查汇总：共 N 处数据库操作，发现 X 个错误、Y 个警告、Z 个建议
```

## 注意事项

- 对GORM Gen自动生成的代码（`dao/` 和 `model/` 目录下带 `.gen.go` 后缀的文件）不需要检查，它们由工具生成且总是正确的
- 重点检查开发者手写的 `logic/` 层代码
- 检查时需要同时参考model定义来验证字段名和表名
- 如果某处代码无法确定是否有问题（比如动态拼接的列名），标注为需人工确认而不是误报
