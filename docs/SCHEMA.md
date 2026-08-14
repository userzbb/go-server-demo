# Omega Server 数据规范文档

## 1. PostgreSQL 表结构（核心数据）

### 1.1 玩家表（players）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | UUID | PRIMARY KEY | 玩家唯一ID |
| username | VARCHAR(32) | UNIQUE NOT NULL | 用户名 |
| password_hash | VARCHAR(255) | NOT NULL | 密码哈希 |
| email | VARCHAR(64) | | 邮箱 |
| nickname | VARCHAR(32) | | 昵称 |
| level | INT | DEFAULT 1 | 等级 |
| exp | BIGINT | DEFAULT 0 | 经验值 |
| gold | BIGINT | DEFAULT 0 | 金币 |
| diamond | BIGINT | DEFAULT 0 | 钻石 |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT NOW() | 更新时间 |
| last_login_at | TIMESTAMP | | 最后登录时间 |

索引：
- idx_players_username ON players(username)
- idx_players_level ON players(level)

### 1.2 玩家资产表（player_assets）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | UUID | PRIMARY KEY | |
| player_id | UUID | NOT NULL | 关联 players.id |
| item_id | VARCHAR(32) | NOT NULL | 物品ID |
| count | INT | DEFAULT 1 | 数量 |
| created_at | TIMESTAMP | DEFAULT NOW() | |
| updated_at | TIMESTAMP | DEFAULT NOW() | |

索引：
- idx_player_assets_player_id ON player_assets(player_id)

### 1.3 房间表（rooms）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | UUID | PRIMARY KEY | 房间ID |
| name | VARCHAR(64) | NOT NULL | 房间名称 |
| owner_id | UUID | NOT NULL | 房主玩家ID |
| max_players | INT | DEFAULT 10 | 最大人数 |
| current_players | INT | DEFAULT 0 | 当前人数 |
| status | VARCHAR(16) | DEFAULT waiting | waiting / playing / closed |
| created_at | TIMESTAMP | DEFAULT NOW() | |
| updated_at | TIMESTAMP | DEFAULT NOW() | |

索引：
- idx_rooms_status ON rooms(status)

### 1.4 房间成员表（room_members）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | UUID | PRIMARY KEY | |
| room_id | UUID | NOT NULL | 关联 rooms.id |
| player_id | UUID | NOT NULL | 关联 players.id |
| joined_at | TIMESTAMP | DEFAULT NOW() | |

索引：
- idx_room_members_room_id ON room_members(room_id)
- idx_room_members_player_id ON room_members(player_id)
- UNIQUE(room_id, player_id)

## 2. MongoDB 集合（灵活数据）

### 2.1 玩家档案（player_profiles）

{
  _id: ObjectId,
  player_id: uuid-string,
  bio: 个人简介,
  avatar: 头像URL,
  achievements: [成就1, 成就2],
  stats: {
    total_games: 0,
    total_wins: 0,
    total_kills: 0
  },
  settings: {
    sound: true,
    vibration: false,
    language: zh-CN
  },
  updated_at: ISODate
}

### 2.2 游戏日志（game_logs）

{
  _id: ObjectId,
  room_id: uuid-string,
  game_type: pvp,
  players: [player1, player2],
  winner: player1,
  duration: 120,
  details: {},
  created_at: ISODate
}

## 3. Redis 数据结构（缓存/实时）

### 3.1 玩家会话

Key: session:{player_id}
Value: {"token": "xxx", "login_at": 1234567890, "gate_addr": "127.0.0.1:8888"}
TTL: 24h

### 3.2 玩家在线状态

Key: online:{player_id}
Value: 1（存在即在线）
TTL: 60s（心跳续期）

### 3.3 排行榜

Key: rank:level
Score: 等级
Value: player_id

### 3.4 房间信息缓存

Key: room:{room_id}
Value: {"name": "xxx", "players": [...], "status": "waiting"}
TTL: 1h

## 4. 枚举定义

### 4.1 房间状态

| 值 | 说明 |
|------|------|
| waiting | 等待中 |
| playing | 游戏中 |
| closed | 已关闭 |

### 4.2 消息ID范围（与 PROTOCOL.md 保持一致）

| 范围 | 用途 |
|------|------|
| 1000-1999 | 登录/认证 |
| 2000-2999 | 房间管理 |
| 3000-3999 | 游戏逻辑 |
| 4000-4999 | 战斗系统 |
| 9000-9999 | 系统消息 |
