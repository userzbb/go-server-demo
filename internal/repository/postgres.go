package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"omega-server/internal/config"
	"omega-server/internal/model"
)

// PostgresPlayerRepository 基于 pgxpool 连接池的玩家仓库实现
type PostgresPlayerRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresPlayerRepository 创建仓库并初始化连接池，同时校验数据库连通性
func NewPostgresPlayerRepository(ctx context.Context, cfg config.Postgres) (*PostgresPlayerRepository, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("解析数据库连接串: %w", err)
	}
	poolCfg.MaxConns = int32(cfg.MaxOpenConn)

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("创建连接池: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("连接 PostgreSQL 失败: %w", err)
	}
	return &PostgresPlayerRepository{pool: pool}, nil
}

// Close 关闭底层连接池
func (r *PostgresPlayerRepository) Close() {
	r.pool.Close()
}

// Create 插入玩家记录，返回填充了 ID 与时间戳的实体
func (r *PostgresPlayerRepository) Create(ctx context.Context, p *model.Player) (*model.Player, error) {
	query := `INSERT INTO players (username, password_hash, email, nickname, level, exp, gold, diamond)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		p.Username, p.PasswordHash, p.Email, p.Nickname, p.Level, p.Exp, p.Gold, p.Diamond,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUsernameExists
		}
		return nil, fmt.Errorf("插入玩家: %w", err)
	}
	return p, nil
}

// GetByUsername 按用户名查询玩家，不存在返回 ErrNotFound
func (r *PostgresPlayerRepository) GetByUsername(ctx context.Context, username string) (*model.Player, error) {
	query := `SELECT id, username, password_hash, COALESCE(email, ''), COALESCE(nickname, ''),
		level, exp, gold, diamond, created_at, updated_at, last_login_at
		FROM players WHERE username = $1`

	var p model.Player
	var email, nickname string
	var lastLoginAt *time.Time
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&p.ID, &p.Username, &p.PasswordHash, &email, &nickname,
		&p.Level, &p.Exp, &p.Gold, &p.Diamond, &p.CreatedAt, &p.UpdatedAt, &lastLoginAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询玩家: %w", err)
	}
	p.Email = email
	p.Nickname = nickname
	p.LastLoginAt = lastLoginAt
	return &p, nil
}

// UpdateLastLogin 更新玩家最后登录时间
func (r *PostgresPlayerRepository) UpdateLastLogin(ctx context.Context, playerID string) error {
	if _, err := r.pool.Exec(ctx,
		`UPDATE players SET last_login_at = NOW() WHERE id = $1`, playerID); err != nil {
		return fmt.Errorf("更新登录时间: %w", err)
	}
	return nil
}
