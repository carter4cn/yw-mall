// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	"mall-common/errorx"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/zeromicro/go-zero/core/logx"
)

type UploadPart struct {
	Filename string
	Reader   io.Reader
	Size     int64
	MIME     string
}

type UploadReviewMediaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadReviewMediaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadReviewMediaLogic {
	return &UploadReviewMediaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadReviewMediaLogic) UploadReviewMedia() (*types.UploadReviewMediaResp, error) {
	return nil, errors.New("use Upload(parts) instead")
}

func (l *UploadReviewMediaLogic) Upload(parts []UploadPart) (*types.UploadReviewMediaResp, error) {
	if len(parts) == 0 {
		return nil, errorx.NewCodeError(errorx.ParamError)
	}
	images, videos := 0, 0
	for _, p := range parts {
		mime := strings.ToLower(p.MIME)
		switch {
		case strings.HasPrefix(mime, "image/"):
			images++
			if p.Size > int64(l.svcCtx.Config.ReviewMedia.MaxImageSizeMB)<<20 {
				return nil, errorx.NewCodeError(errorx.ReviewLimitExceeded)
			}
		case strings.HasPrefix(mime, "video/"):
			videos++
			if p.Size > int64(l.svcCtx.Config.ReviewMedia.MaxVideoSizeMB)<<20 {
				return nil, errorx.NewCodeError(errorx.ReviewLimitExceeded)
			}
		default:
			return nil, errorx.NewCodeError(errorx.ReviewMediaInvalid)
		}
	}
	if images > l.svcCtx.Config.ReviewMedia.MaxImages || videos > 1 {
		return nil, errorx.NewCodeError(errorx.ReviewLimitExceeded)
	}

	bucket := l.svcCtx.Config.ReviewMedia.Bucket
	if exists, _ := l.svcCtx.Minio.BucketExists(l.ctx, bucket); !exists {
		_ = l.svcCtx.Minio.MakeBucket(l.ctx, bucket, minio.MakeBucketOptions{})
	}
	out := types.UploadReviewMediaResp{Media: make([]types.ReviewMediaItem, 0, len(parts))}
	for i, p := range parts {
		ext := strings.ToLower(filepath.Ext(p.Filename))
		objKey := fmt.Sprintf("%d/%s%s", time.Now().Unix(), uuid.NewString(), ext)
		if _, err := l.svcCtx.Minio.PutObject(l.ctx, bucket, objKey, p.Reader, p.Size, minio.PutObjectOptions{ContentType: p.MIME}); err != nil {
			return nil, err
		}
		mediaType := int32(1)
		if strings.HasPrefix(strings.ToLower(p.MIME), "video/") {
			mediaType = 2
		}
		out.Media = append(out.Media, types.ReviewMediaItem{
			Type: mediaType,
			Url:  fmt.Sprintf("http://%s/%s/%s", l.svcCtx.Config.MinIO.Endpoint, bucket, objKey),
			Sort: int32(i),
		})
	}
	return &out, nil
}
