package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"bit-labs.cn/owl/provider/redis"
	"bit-labs.cn/owl/provider/router"
	"bit-labs.cn/owl-metrics/app/model"
	"bit-labs.cn/owl-metrics/app/repository"
	validatorv10 "github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type CreateMetricsAppReq struct {
	Name           string   `json:"name" validate:"required,max=120" label:"应用名称"`
	Description    string   `json:"description" validate:"omitempty,max=500" label:"描述"`
	AllowedOrigins []string `json:"allowedOrigins" validate:"omitempty,dive,max=256" label:"允许来源"`
}

type UpdateMetricsAppReq struct {
	ID             uint     `json:"id,string" validate:"required,gt=0" label:"ID"`
	Name           string   `json:"name" validate:"required,max=120" label:"应用名称"`
	Description    string   `json:"description" validate:"omitempty,max=500" label:"描述"`
	AllowedOrigins []string `json:"allowedOrigins" validate:"omitempty,dive,max=256" label:"允许来源"`
	Status         int      `json:"status" validate:"omitempty,oneof=1 2" label:"状态"`
}

type MetricsAppResp struct {
	ID             uint     `json:"id,string"`
	Name           string   `json:"name"`
	AppKey         string   `json:"appKey"`
	AppSecret      string   `json:"appSecret,omitempty"`
	AllowedOrigins []string `json:"allowedOrigins"`
	Status         int      `json:"status"`
	Description    string   `json:"description"`
	CreatedAt      int64    `json:"createdAt"`
	UpdatedAt      int64    `json:"updatedAt"`
}

type MetricsAppService struct {
	repo     repository.MetricsAppRepositoryInterface
	locker   redis.LockerFactory
	validate *validatorv10.Validate
}

// NewMetricsAppService 创建 MetricsAppService 实例。
func NewMetricsAppService(
	repo repository.MetricsAppRepositoryInterface,
	locker redis.LockerFactory,
	validate *validatorv10.Validate,
) *MetricsAppService {
	return &MetricsAppService{repo: repo, locker: locker, validate: validate}
}

// Create 创建指标应用。
func (s *MetricsAppService) Create(ctx context.Context, req *CreateMetricsAppReq) (*MetricsAppResp, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}
	l := s.locker.New()
	if err := l.Lock("metrics:app:create"); err != nil {
		return nil, err
	}
	defer l.Unlock()

	appKey, err := genToken(16)
	if err != nil {
		return nil, err
	}
	appSecret, err := genToken(32)
	if err != nil {
		return nil, err
	}
	m := &model.MetricsApp{
		Name:           strings.TrimSpace(req.Name),
		AppKey:         appKey,
		AppSecret:      appSecret,
		AllowedOrigins: encodeOrigins(req.AllowedOrigins),
		Status:         model.MetricsAppStatusActive,
		Description:    strings.TrimSpace(req.Description),
	}
	if err := s.repo.WithContext(ctx).Save(m); err != nil {
		return nil, err
	}
	return toAppResp(m, true), nil
}

// Update 更新指标应用。
func (s *MetricsAppService) Update(ctx context.Context, req *UpdateMetricsAppReq) error {
	if err := s.validate.Struct(req); err != nil {
		return err
	}
	l := s.locker.New()
	if err := l.Lock("metrics:app:update:" + cast.ToString(req.ID)); err != nil {
		return err
	}
	defer l.Unlock()

	cur, err := s.repo.WithContext(ctx).Detail(req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return AppNotFound()
		}
		return err
	}
	cur.Name = strings.TrimSpace(req.Name)
	cur.Description = strings.TrimSpace(req.Description)
	cur.AllowedOrigins = encodeOrigins(req.AllowedOrigins)
	if req.Status == model.MetricsAppStatusActive || req.Status == model.MetricsAppStatusDisabled {
		cur.Status = req.Status
	}
	return s.repo.WithContext(ctx).Save(cur)
}

// Delete 删除指标应用。
func (s *MetricsAppService) Delete(ctx context.Context, id uint) error {
	l := s.locker.New()
	if err := l.Lock("metrics:app:delete:" + cast.ToString(id)); err != nil {
		return err
	}
	defer l.Unlock()
	return s.repo.WithContext(ctx).Delete(id)
}

// Detail 获取详情。
func (s *MetricsAppService) Detail(ctx context.Context, id uint) (*MetricsAppResp, error) {
	m, err := s.repo.WithContext(ctx).Detail(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, AppNotFound()
		}
		return nil, err
	}
	return toAppResp(m, true), nil
}

// Retrieve 分页列表。
func (s *MetricsAppService) Retrieve(ctx context.Context, req *router.PageReq) (int64, []MetricsAppResp, error) {
	count, list, err := s.repo.WithContext(ctx).Retrieve(req.Page, req.PageSize, func(tx *gorm.DB) {
		tx.Order("created_at desc")
	})
	if err != nil {
		return 0, nil, err
	}
	out := make([]MetricsAppResp, 0, len(list))
	for i := range list {
		out = append(out, *toAppResp(&list[i], false))
	}
	return count, out, nil
}

// FindByAppKey 按 AppKey 查询（采集用）。
func (s *MetricsAppService) FindByAppKey(ctx context.Context, appKey string) (*model.MetricsApp, error) {
	m, err := s.repo.WithContext(ctx).FindByAppKey(strings.TrimSpace(appKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, AppNotFound()
		}
		return nil, err
	}
	if m.Status != model.MetricsAppStatusActive {
		return nil, AppDisabled()
	}
	return m, nil
}

// RegenerateSecret 重新生成 AppSecret。
func (s *MetricsAppService) RegenerateSecret(ctx context.Context, id uint) (string, error) {
	l := s.locker.New()
	if err := l.Lock("metrics:app:secret:" + cast.ToString(id)); err != nil {
		return "", err
	}
	defer l.Unlock()

	cur, err := s.repo.WithContext(ctx).Detail(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", AppNotFound()
		}
		return "", err
	}
	secret, err := genToken(32)
	if err != nil {
		return "", err
	}
	cur.AppSecret = secret
	if err := s.repo.WithContext(ctx).Save(cur); err != nil {
		return "", err
	}
	return secret, nil
}

func toAppResp(m *model.MetricsApp, includeSecret bool) *MetricsAppResp {
	resp := &MetricsAppResp{
		ID:             m.ID,
		Name:           m.Name,
		AppKey:         m.AppKey,
		AllowedOrigins: decodeOrigins(m.AllowedOrigins),
		Status:         m.Status,
		Description:    m.Description,
	}
	if m.CreatedAt != nil {
		resp.CreatedAt = m.CreatedAt.Unix()
	}
	if m.UpdatedAt != nil {
		resp.UpdatedAt = m.UpdatedAt.Unix()
	}
	if includeSecret {
		resp.AppSecret = m.AppSecret
	}
	return resp
}

func encodeOrigins(origins []string) string {
	if len(origins) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(origins)
	return string(b)
}

func decodeOrigins(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var out []string
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func genToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CopyAppResp 复制响应（隐藏 secret）。
func CopyAppResp(src *MetricsAppResp) MetricsAppResp {
	var dst MetricsAppResp
	_ = copier.Copy(&dst, src)
	dst.AppSecret = ""
	return dst
}
