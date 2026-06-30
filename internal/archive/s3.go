package archive

import (
	"aws_fin_ops/internal/utilities"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	// s3ArchiveComponent 是 S3 归档功能在日志中使用的组件名称标识。
	s3ArchiveComponent = "S3Archive"

	// s3ArchiveTimeout 是 S3 单次操作的默认超时时间。
	s3ArchiveTimeout = 2 * time.Minute

	// defaultArchiveRegion 是未配置 S3_ARCHIVE_REGION 时的默认区域。
	defaultArchiveRegion = "ap-east-1"
)

// Config 包含 S3 Glacier 归档所需的配置项。
// 所有字段均从环境变量读取，未设置时归档功能自动禁用。
type Config struct {
	// Bucket 用于存储归档文件的 S3 Bucket 名称。
	// 为空时表示不启用归档功能。
	Bucket string

	// Region Bucket 所在的 AWS 区域。
	Region string

	// StorageClass 归档使用的存储类型，默认 GLACIER_IR（Glacier Instant Retrieval）。
	// 可选值：STANDARD_IA、GLACIER_IR、GLACIER、DEEP_ARCHIVE
	StorageClass string

	// Prefix 归档对象在 Bucket 中的前缀（类似子目录）。
	Prefix string
}

// LoadConfig 从环境变量加载 S3 Glacier 归档配置。
//
// 返回：
//   - *Config : 加载后的配置，未配置 Bucket 时返回 nil
func LoadConfig() *Config {
	bucket := os.Getenv("S3_ARCHIVE_BUCKET")
	if bucket == "" {
		return nil
	}

	cfg := &Config{
		Bucket:       bucket,
		Region:       utilities.AWSRegion(defaultArchiveRegion),
		StorageClass: "GLACIER_IR",
		Prefix:       "finops-reports/",
	}

	if r := os.Getenv("S3_ARCHIVE_REGION"); r != "" {
		cfg.Region = r
	}
	if sc := os.Getenv("S3_ARCHIVE_STORAGE_CLASS"); sc != "" {
		cfg.StorageClass = sc
	}
	if p := os.Getenv("S3_ARCHIVE_PREFIX"); p != "" {
		cfg.Prefix = p
	}

	return cfg
}

// ArchiveFiles 将本地文件列表归档到 S3 Glacier 存储。
// 如果指定的 Bucket 不存在，会自动创建。
// 单个文件上传失败不影响其他文件的归档。
//
// 参数：
//   - cfg   : 归档配置
//   - files : 本地文件路径列表
//
// 返回：
//   - []string : 成功归档的 S3 对象 Key 列表
//   - error    : 整体流程失败时返回错误（如 Bucket 创建失败）
func ArchiveFiles(cfg *Config, files []string) ([]string, error) {
	start := time.Now()
	utilities.LogStart(s3ArchiveComponent, "ArchiveFiles")

	if cfg == nil || cfg.Bucket == "" {
		utilities.LogProgress(s3ArchiveComponent, "ArchiveFiles",
			"S3 归档未配置，跳过",
		)
		return nil, nil
	}

	if len(files) == 0 {
		utilities.LogProgress(s3ArchiveComponent, "ArchiveFiles",
			"没有需要归档的文件",
		)
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), s3ArchiveTimeout)
	defer cancel()

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
	)
	if err != nil {
		utilities.LogError(s3ArchiveComponent, "ArchiveFiles", err, time.Since(start), "step=load_config")
		return nil, fmt.Errorf("加载 AWS 配置失败: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)

	if err := ensureBucketExists(ctx, client, cfg); err != nil {
		utilities.LogError(s3ArchiveComponent, "ArchiveFiles", err, time.Since(start),
			fmt.Sprintf("bucket=%s", cfg.Bucket),
			"step=ensure_bucket",
		)
		return nil, fmt.Errorf("确保 Bucket 存在失败: %w", err)
	}

	var archived []string
	for _, filePath := range files {
		key, err := uploadFileToGlacier(ctx, client, cfg, filePath)
		if err != nil {
			utilities.LogWarn(s3ArchiveComponent, "ArchiveFiles",
				fmt.Sprintf("文件 %s 上传失败: %v", filepath.Base(filePath), err),
				time.Since(start),
			)
			continue
		}
		archived = append(archived, key)
	}

	utilities.LogSuccess(s3ArchiveComponent, "ArchiveFiles", time.Since(start),
		fmt.Sprintf("bucket=%s", cfg.Bucket),
		fmt.Sprintf("total=%d", len(files)),
		fmt.Sprintf("archived=%d", len(archived)),
		fmt.Sprintf("storage_class=%s", cfg.StorageClass),
	)
	return archived, nil
}

// ensureBucketExists 检查指定的 S3 Bucket 是否存在，不存在则创建。
// 创建时会自动配置公共访问阻止和默认服务器端加密。
//
// 参数：
//   - ctx    : 上下文
//   - client : S3 客户端
//   - cfg    : 归档配置
//
// 返回：
//   - error : 操作失败时返回错误
func ensureBucketExists(ctx context.Context, client *s3.Client, cfg *Config) error {
	start := time.Now()

	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &cfg.Bucket,
	})
	if err == nil {
		utilities.LogProgress(s3ArchiveComponent, "ensureBucketExists",
			fmt.Sprintf("Bucket 已存在: %s", cfg.Bucket),
		)
		return nil
	}

	utilities.LogProgress(s3ArchiveComponent, "ensureBucketExists",
		fmt.Sprintf("Bucket %s 不存在，正在创建", cfg.Bucket),
	)

	createInput := &s3.CreateBucketInput{
		Bucket:          &cfg.Bucket,
		ObjectOwnership: types.ObjectOwnershipBucketOwnerPreferred,
	}

	if cfg.Region != "us-east-1" {
		createInput.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(cfg.Region),
		}
	}

	_, err = client.CreateBucket(ctx, createInput)
	if err != nil {
		return fmt.Errorf("创建 Bucket 失败 (bucket=%s, region=%s): %w",
			cfg.Bucket, cfg.Region, err)
	}

	publicAccess := &types.PublicAccessBlockConfiguration{
		BlockPublicAcls:       awsBool(true),
		IgnorePublicAcls:      awsBool(true),
		BlockPublicPolicy:     awsBool(true),
		RestrictPublicBuckets: awsBool(true),
	}

	_, _ = client.PutPublicAccessBlock(ctx, &s3.PutPublicAccessBlockInput{
		Bucket:                         &cfg.Bucket,
		PublicAccessBlockConfiguration: publicAccess,
	})

	_, _ = client.PutBucketEncryption(ctx, &s3.PutBucketEncryptionInput{
		Bucket: &cfg.Bucket,
		ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
			Rules: []types.ServerSideEncryptionRule{
				{
					ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{
						SSEAlgorithm: types.ServerSideEncryptionAes256,
					},
				},
			},
		},
	})

	utilities.LogSuccess(s3ArchiveComponent, "ensureBucketExists", time.Since(start),
		fmt.Sprintf("bucket=%s", cfg.Bucket),
		fmt.Sprintf("region=%s", cfg.Region),
	)
	return nil
}

// uploadFileToGlacier 将单个文件上传到 S3 并使用指定的 Glacier 存储类。
// 对象 Key 格式：{prefix}/{YYYY-MM-DD}/{filename}
//
// 参数：
//   - ctx      : 上下文
//   - client   : S3 客户端
//   - cfg      : 归档配置
//   - filePath : 本地文件路径
//
// 返回：
//   - string : 上传成功的 S3 对象 Key
//   - error  : 上传失败时返回错误
func uploadFileToGlacier(ctx context.Context, client *s3.Client, cfg *Config, filePath string) (string, error) {
	start := time.Now()
	fileName := filepath.Base(filePath)

	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("%s%s/%s", cfg.Prefix, today, fileName)

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("获取文件信息失败: %w", err)
	}

	contentType := detectContentType(fileName)

	contentLength := fileInfo.Size()

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        &cfg.Bucket,
		Key:           &key,
		Body:          file,
		ContentLength: &contentLength,
		ContentType:   &contentType,
		StorageClass:  types.StorageClass(cfg.StorageClass),
	})
	if err != nil {
		return "", fmt.Errorf("PutObject 失败 (key=%s): %w", key, err)
	}

	utilities.LogProgress(s3ArchiveComponent, "uploadFileToGlacier",
		fmt.Sprintf("已归档: s3://%s/%s", cfg.Bucket, key),
		fmt.Sprintf("size=%d bytes", fileInfo.Size()),
		fmt.Sprintf("elapsed=%s", time.Since(start).Round(time.Millisecond)),
	)
	return key, nil
}

// detectContentType 根据文件扩展名推断 Content-Type。
//
// 参数：
//   - fileName : 文件名
//
// 返回：
//   - string : MIME 类型
func detectContentType(fileName string) string {
	ext := filepath.Ext(fileName)
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".csv":
		return "text/csv; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

// awsBool 返回 bool 指针，用于 AWS SDK API 中需要 *bool 的字段。
func awsBool(v bool) *bool {
	return &v
}
