package aws

// RightsizingType 定义资源调整建议的类型。
type RightsizingType string

const (
	// RightsizingTerminate 建议终止该实例（资源完全空闲或极度低利用率）。
	RightsizingTerminate RightsizingType = "TERMINATE"

	// RightsizingModify 建议变更实例规格（降配或换代以降低成本）。
	RightsizingModify RightsizingType = "MODIFY"
)

// FindingReasonCode 定义触发调整建议的具体原因代码。
type FindingReasonCode string

const (
	// ReasonCPUOverProvisioned CPU 过度配置，实际使用率远低于分配。
	ReasonCPUOverProvisioned FindingReasonCode = "CPU_OVER_PROVISIONED"

	// ReasonCPUUnderProvisioned CPU 配置不足，实际使用率接近或超过上限。
	ReasonCPUUnderProvisioned FindingReasonCode = "CPU_UNDER_PROVISIONED"

	// ReasonMemoryOverProvisioned 内存过度配置。
	ReasonMemoryOverProvisioned FindingReasonCode = "MEMORY_OVER_PROVISIONED"

	// ReasonMemoryUnderProvisioned 内存配置不足。
	ReasonMemoryUnderProvisioned FindingReasonCode = "MEMORY_UNDER_PROVISIONED"

	// ReasonDiskIOPSOverProvisioned 磁盘 IOPS 过度配置。
	ReasonDiskIOPSOverProvisioned FindingReasonCode = "DISK_IOPS_OVER_PROVISIONED"

	// ReasonDiskIOPSUnderProvisioned 磁盘 IOPS 配置不足。
	ReasonDiskIOPSUnderProvisioned FindingReasonCode = "DISK_IOPS_UNDER_PROVISIONED"

	// ReasonDiskThroughputOverProvisioned 磁盘吞吐量过度配置。
	ReasonDiskThroughputOverProvisioned FindingReasonCode = "DISK_THROUGHPUT_OVER_PROVISIONED"

	// ReasonDiskThroughputUnderProvisioned 磁盘吞吐量配置不足。
	ReasonDiskThroughputUnderProvisioned FindingReasonCode = "DISK_THROUGHPUT_UNDER_PROVISIONED"

	// ReasonNetworkBandwidthOverProvisioned 网络带宽过度配置。
	ReasonNetworkBandwidthOverProvisioned FindingReasonCode = "NETWORK_BANDWIDTH_OVER_PROVISIONED"

	// ReasonNetworkBandwidthUnderProvisioned 网络带宽配置不足。
	ReasonNetworkBandwidthUnderProvisioned FindingReasonCode = "NETWORK_BANDWIDTH_UNDER_PROVISIONED"

	// ReasonNetworkPPSOverProvisioned 网络 PPS（每秒包数）过度配置。
	ReasonNetworkPPSOverProvisioned FindingReasonCode = "NETWORK_PPS_OVER_PROVISIONED"

	// ReasonNetworkPPSUnderProvisioned 网络 PPS 配置不足。
	ReasonNetworkPPSUnderProvisioned FindingReasonCode = "NETWORK_PPS_UNDER_PROVISIONED"

	// ReasonEBSIOPSOverProvisioned EBS IOPS 过度配置。
	ReasonEBSIOPSOverProvisioned FindingReasonCode = "EBS_IOPS_OVER_PROVISIONED"

	// ReasonEBSIOPSUnderProvisioned EBS IOPS 配置不足。
	ReasonEBSIOPSUnderProvisioned FindingReasonCode = "EBS_IOPS_UNDER_PROVISIONED"

	// ReasonEBSThroughputOverProvisioned EBS 吞吐量过度配置。
	ReasonEBSThroughputOverProvisioned FindingReasonCode = "EBS_THROUGHPUT_OVER_PROVISIONED"

	// ReasonEBSThroughputUnderProvisioned EBS 吞吐量配置不足。
	ReasonEBSThroughputUnderProvisioned FindingReasonCode = "EBS_THROUGHPUT_UNDER_PROVISIONED"
)

// PaymentOption 定义建议方案的付费方式。
type PaymentOption string

const (
	// PaymentNoUpfront 无预付（全部按月付费）。
	PaymentNoUpfront PaymentOption = "NO_UPFRONT"

	// PaymentPartialUpfront 部分预付（预付一部分，其余按月）。
	PaymentPartialUpfront PaymentOption = "PARTIAL_UPFRONT"

	// PaymentAllUpfront 全部预付（一次性支付，月费为零）。
	PaymentAllUpfront PaymentOption = "ALL_UPFRONT"
)

// EC2ResourceDetails 描述 EC2 实例的硬件和定价属性。
//
// 字段：
//   - HourlyOnDemandRate : 按需小时费率
//   - InstanceType       : 实例类型（如 m5.xlarge）
//   - Memory             : 内存大小（如 "16 GiB"）
//   - NetworkPerformance : 网络性能等级（如 "Up to 10 Gbps"）
//   - Platform           : 操作系统平台（如 "Linux/UNIX"）
//   - Region             : 所在区域
//   - Sku                : SKU 标识
//   - Storage            : 存储配置（如 "EBS only"）
//   - Vcpu               : vCPU 核数
type EC2ResourceDetails struct {
	HourlyOnDemandRate string `json:"hourly_on_demand_rate"`
	InstanceType       string `json:"instance_type"`
	Memory             string `json:"memory"`
	NetworkPerformance string `json:"network_performance"`
	Platform           string `json:"platform"`
	Region             string `json:"region"`
	Sku                string `json:"sku"`
	Storage            string `json:"storage"`
	Vcpu               string `json:"vcpu"`
}

// EC2ResourceUtilization 描述 EC2 实例的资源利用率指标。
//
// 字段：
//   - MaxCpuUtilizationPercentage    : CPU 最大使用率
//   - MaxMemoryUtilizationPercentage : 内存最大使用率
//   - MaxStorageUtilizationPercentage: 存储最大使用率
//   - EBSResourceUtilization         : EBS 子系统利用率
//   - DiskResourceUtilization        : 本地磁盘子系统利用率
//   - NetworkResourceUtilization     : 网络子系统利用率
type EC2ResourceUtilization struct {
	MaxCpuUtilizationPercentage     string                    `json:"max_cpu_utilization_percentage"`
	MaxMemoryUtilizationPercentage  string                    `json:"max_memory_utilization_percentage"`
	MaxStorageUtilizationPercentage string                    `json:"max_storage_utilization_percentage"`
	EBSResourceUtilization          *EBSResourceUtilization   `json:"ebs_resource_utilization,omitempty"`
	DiskResourceUtilization         *DiskResourceUtilization  `json:"disk_resource_utilization,omitempty"`
	NetworkResourceUtilization      *NetworkResourceUtilization `json:"network_resource_utilization,omitempty"`
}

// EBSResourceUtilization 描述 EBS 卷的利用率指标。
//
// 字段：
//   - EbsReadBytesPerSecond  : 每秒读取字节数
//   - EbsReadOpsPerSecond    : 每秒读取操作数
//   - EbsWriteBytesPerSecond : 每秒写入字节数
//   - EbsWriteOpsPerSecond   : 每秒写入操作数
type EBSResourceUtilization struct {
	EbsReadBytesPerSecond  string `json:"ebs_read_bytes_per_second"`
	EbsReadOpsPerSecond    string `json:"ebs_read_ops_per_second"`
	EbsWriteBytesPerSecond string `json:"ebs_write_bytes_per_second"`
	EbsWriteOpsPerSecond   string `json:"ebs_write_ops_per_second"`
}

// DiskResourceUtilization 描述本地磁盘的利用率指标。
//
// 字段：
//   - DiskReadBytesPerSecond  : 每秒读取字节数
//   - DiskReadOpsPerSecond    : 每秒读取操作数
//   - DiskWriteBytesPerSecond : 每秒写入字节数
//   - DiskWriteOpsPerSecond   : 每秒写入操作数
type DiskResourceUtilization struct {
	DiskReadBytesPerSecond  string `json:"disk_read_bytes_per_second"`
	DiskReadOpsPerSecond    string `json:"disk_read_ops_per_second"`
	DiskWriteBytesPerSecond string `json:"disk_write_bytes_per_second"`
	DiskWriteOpsPerSecond   string `json:"disk_write_ops_per_second"`
}

// NetworkResourceUtilization 描述网络的利用率指标。
//
// 字段：
//   - NetworkInBytesPerSecond  : 每秒入站字节数
//   - NetworkOutBytesPerSecond : 每秒出站字节数
//   - NetworkPacketsInPerSecond  : 每秒入站数据包数
//   - NetworkPacketsOutPerSecond : 每秒出站数据包数
type NetworkResourceUtilization struct {
	NetworkInBytesPerSecond    string `json:"network_in_bytes_per_second"`
	NetworkOutBytesPerSecond   string `json:"network_out_bytes_per_second"`
	NetworkPacketsInPerSecond  string `json:"network_packets_in_per_second"`
	NetworkPacketsOutPerSecond string `json:"network_packets_out_per_second"`
}

// ResourceDetails 封装不同 AWS 服务的资源详情。
// 目前仅支持 EC2，后续可扩展 RDS、Lambda 等。
//
// 字段：
//   - EC2ResourceDetails : EC2 实例的硬件和定价属性
type ResourceDetails struct {
	EC2ResourceDetails *EC2ResourceDetails `json:"ec2_resource_details,omitempty"`
}

// ResourceUtilization 封装不同 AWS 服务的资源利用率。
//
// 字段：
//   - EC2ResourceUtilization : EC2 实例的利用率指标
type ResourceUtilization struct {
	EC2ResourceUtilization *EC2ResourceUtilization `json:"ec2_resource_utilization,omitempty"`
}

// CurrentInstance 描述当前运行中的实例信息（被评估的实例）。
//
// 字段：
//   - ResourceID          : AWS 资源 ID（如 i-0abc123def456）
//   - InstanceName        : 实例名称标签
//   - Tags                : 资源标签列表
//   - ResourceDetails     : 实例硬件和定价属性
//   - ResourceUtilization : 实例资源利用率
//   - MonthlyCost         : 当前月度费用
//   - CurrencyCode        : 货币代码
//   - OnDemandHoursInLookbackPeriod : 回溯期内的按需小时数
//   - ReservationCoveredHoursInLookbackPeriod : 回溯期内 RI 覆盖的小时数
//   - SavingsPlansCoveredHoursInLookbackPeriod : 回溯期内 SP 覆盖的小时数
//   - TotalRunningHoursInLookbackPeriod : 回溯期内总运行小时数
type CurrentInstance struct {
	ResourceID                                string               `json:"resource_id"`
	InstanceName                              string               `json:"instance_name"`
	Tags                                      []TagValue           `json:"tags,omitempty"`
	ResourceDetails                           ResourceDetails      `json:"resource_details"`
	ResourceUtilization                       ResourceUtilization  `json:"resource_utilization"`
	MonthlyCost                               string               `json:"monthly_cost"`
	CurrencyCode                              string               `json:"currency_code"`
	OnDemandHoursInLookbackPeriod             string               `json:"on_demand_hours_in_lookback_period"`
	ReservationCoveredHoursInLookbackPeriod    string               `json:"reservation_covered_hours_in_lookback_period"`
	SavingsPlansCoveredHoursInLookbackPeriod   string               `json:"savings_plans_covered_hours_in_lookback_period"`
	TotalRunningHoursInLookbackPeriod          string               `json:"total_running_hours_in_lookback_period"`
}

// TagValue 表示资源标签的键值对。
//
// 字段：
//   - Key   : 标签键
//   - Value : 标签值
type TagValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// TargetInstance 描述建议变更后的目标实例配置。
//
// 字段：
//   - EstimatedMonthlyCost    : 变更后预估的月度费用
//   - EstimatedMonthlySavings : 变更后预估的月度节省金额
//   - CurrencyCode            : 货币代码
//   - DefaultTargetInstance   : 是否为默认推荐的目标实例
//   - ResourceDetails         : 目标实例的硬件和定价属性
//   - ExpectedResourceUtilization : 变更后预期的资源利用率
type TargetInstance struct {
	EstimatedMonthlyCost        string              `json:"estimated_monthly_cost"`
	EstimatedMonthlySavings     string              `json:"estimated_monthly_savings"`
	CurrencyCode                string              `json:"currency_code"`
	DefaultTargetInstance       bool                `json:"default_target_instance"`
	ResourceDetails             ResourceDetails     `json:"resource_details"`
	ExpectedResourceUtilization ResourceUtilization `json:"expected_resource_utilization"`
}

// ModifyRecommendationDetail 描述"变更实例规格"建议的详细信息。
//
// 字段：
//   - TargetInstances : 建议的目标实例列表（可能有多个候选方案）
type ModifyRecommendationDetail struct {
	TargetInstances []TargetInstance `json:"target_instances"`
}

// TerminateRecommendationDetail 描述"终止实例"建议的详细信息。
//
// 字段：
//   - EstimatedMonthlySavings : 终止后预估的月度节省金额
//   - CurrencyCode            : 货币代码
type TerminateRecommendationDetail struct {
	EstimatedMonthlySavings string `json:"estimated_monthly_savings"`
	CurrencyCode            string `json:"currency_code"`
}

// RightsizingRecommendation 表示单条资源优化建议。
//
// 字段：
//   - AccountId                      : 资源所属的 AWS 账户 ID
//   - CurrentInstance                : 当前实例信息
//   - RightsizingType                : 建议类型（TERMINATE / MODIFY）
//   - FindingReasonCodes             : 触发建议的原因代码列表
//   - ModifyRecommendationDetail     : 变更建议详情（仅 MODIFY 类型）
//   - TerminateRecommendationDetail  : 终止建议详情（仅 TERMINATE 类型）
type RightsizingRecommendation struct {
	AccountId                     string                          `json:"account_id"`
	CurrentInstance               CurrentInstance                 `json:"current_instance"`
	RightsizingType               RightsizingType                 `json:"rightsizing_type"`
	FindingReasonCodes            []FindingReasonCode             `json:"finding_reason_codes"`
	ModifyRecommendationDetail    *ModifyRecommendationDetail     `json:"modify_recommendation_detail,omitempty"`
	TerminateRecommendationDetail *TerminateRecommendationDetail  `json:"terminate_recommendation_detail,omitempty"`
}

// RightsizingRecommendationSummary 表示调整建议的汇总统计信息。
//
// 字段：
//   - TotalRecommendationCount         : 建议总数
//   - EstimatedTotalMonthlySavingsAmount: 预估总月度节省金额
//   - SavingsCurrencyCode               : 节省金额的货币代码
//   - SavingsPercentage                 : 节省百分比
type RightsizingRecommendationSummary struct {
	TotalRecommendationCount          string `json:"total_recommendation_count"`
	EstimatedTotalMonthlySavingsAmount string `json:"estimated_total_monthly_savings_amount"`
	SavingsCurrencyCode               string `json:"savings_currency_code"`
	SavingsPercentage                 string `json:"savings_percentage"`
}

// RightsizingRecommendationConfiguration 控制获取建议时的配置参数。
//
// 字段：
//   - BenefitsConsidered : 是否将 RI/SP 节省计入建议收益中
//   - RecommendationTarget : 建议目标（SAME_INSTANCE_FAMILY / CROSS_INSTANCE_FAMILY）
type RightsizingRecommendationConfiguration struct {
	BenefitsConsidered   bool   `json:"benefits_considered"`
	RecommendationTarget string `json:"recommendation_target"`
}

// GetRightsizingRecommendationRequest 是 GetRightsizingRecommendation API 的请求结构体。
//
// 字段：
//   - Service        : 要获取建议的 AWS 服务（如 "AmazonEC2"，必填）
//   - Filter         : 过滤表达式（可选）
//   - Configuration  : 建议配置（可选）
//   - PageSize       : 每页条数（可选）
//   - NextPageToken  : 分页令牌（可选）
type GetRightsizingRecommendationRequest struct {
	Service       string                                   `json:"service"`
	Filter        *Expression                              `json:"filter,omitempty"`
	Configuration *RightsizingRecommendationConfiguration  `json:"configuration,omitempty"`
	PageSize      int                                      `json:"page_size,omitempty"`
	NextPageToken string                                   `json:"next_page_token,omitempty"`
}

// GetRightsizingRecommendationResponse 是 GetRightsizingRecommendation API 的响应结构体。
//
// 字段：
//   - Metadata        : 请求元数据（含回溯天数、生成时间戳等）
//   - Summary         : 建议汇总统计
//   - Recommendations : 具体建议列表
//   - Configuration   : 使用的配置参数
//   - NextPageToken   : 下一页令牌
type GetRightsizingRecommendationResponse struct {
	Metadata        *RightsizingRecommendationMetadata       `json:"metadata,omitempty"`
	Summary         *RightsizingRecommendationSummary        `json:"summary,omitempty"`
	Recommendations []RightsizingRecommendation              `json:"recommendations"`
	Configuration   *RightsizingRecommendationConfiguration  `json:"configuration,omitempty"`
	NextPageToken   string                                   `json:"next_page_token,omitempty"`
}

// RightsizingRecommendationMetadata 包含建议查询的元数据信息。
//
// 字段：
//   - RecommendationId    : 建议请求的唯一标识
//   - GenerationTimestamp : 建议生成的 UTC 时间戳
//   - LookbackPeriodInDays: 分析回溯天数（默认 14 天）
//   - AdditionalMetadata  : 其他扩展元数据
type RightsizingRecommendationMetadata struct {
	RecommendationId     string `json:"recommendation_id"`
	GenerationTimestamp  string `json:"generation_timestamp"`
	LookbackPeriodInDays string `json:"lookback_period_in_days"`
	AdditionalMetadata   string `json:"additional_metadata,omitempty"`
}
