package safety

import (
	"regexp"
	"strings"

	"github.com/vucongthanh92/courier/agent-gateway/config"
	"github.com/vucongthanh92/courier/agent-gateway/helper/constants"
	"github.com/vucongthanh92/courier/agent-gateway/internal/domain/models"
)

type Guardrail struct {
	enabled           bool
	blockedCategories map[string]struct{}
	rules             []rule
}

type rule struct {
	category string
	reason   string
	patterns []*regexp.Regexp
}

func NewGuardrail(cfg config.SafetyConfig) *Guardrail {
	blocked := make(map[string]struct{}, len(cfg.BlockedCategories))
	for _, category := range cfg.BlockedCategories {
		blocked[strings.TrimSpace(category)] = struct{}{}
	}
	return &Guardrail{
		enabled:           cfg.GuardrailEnabled,
		blockedCategories: blocked,
		rules:             defaultRules(),
	}
}

func (g *Guardrail) Evaluate(text string) models.SafetyEvaluationResult {
	normalized := normalize(text)
	if !g.enabled || normalized == "" {
		return models.SafetyEvaluationResult{Decision: constants.SafetyDecisionAllow}
	}

	for _, currentRule := range g.rules {
		if !g.isBlocked(currentRule.category) {
			continue
		}
		for _, pattern := range currentRule.patterns {
			if pattern.MatchString(normalized) {
				return models.SafetyEvaluationResult{
					Decision:    constants.SafetyDecisionBlock,
					Category:    currentRule.category,
					Reasons:     []string{currentRule.reason},
					UserMessage: blockedUserMessage(currentRule.category),
				}
			}
		}
	}

	return models.SafetyEvaluationResult{Decision: constants.SafetyDecisionAllow}
}

func (g *Guardrail) isBlocked(category string) bool {
	_, ok := g.blockedCategories[category]
	return ok
}

func defaultRules() []rule {
	return []rule{
		{
			category: constants.SafetyCategorySecrets,
			reason:   "request appears to ask for secrets, credentials, tokens, keys, or bypass instructions",
			patterns: compilePatterns(
				`(?i)\b(api[_ -]?key|secret[_ -]?key|access[_ -]?token|refresh[_ -]?token|private[_ -]?key|password|credential|jwt|ssh[_ -]?key)\b`,
				`(?i)\b(show|reveal|print|dump|extract|steal|get|leak|bypass)\b.{0,80}\b(secret|token|key|password|credential|jwt)\b`,
				`(?i)-----begin (rsa |ec |openssh |private )?private key-----`,
			),
		},
		{
			category: constants.SafetyCategoryCyberAbuse,
			reason:   "request appears to ask for cyber abuse, malware, phishing, exploitation, or credential theft",
			patterns: compilePatterns(
				`(?i)\b(malware|ransomware|keylogger|phishing|credential theft|steal credentials|reverse shell|botnet)\b`,
				`(?i)\b(exploit|hack|bypass auth|sql injection|xss payload|privilege escalation)\b`,
				`(?i)\b(write|generate|create|build)\b.{0,80}\b(phishing|malware|ransomware|keylogger|exploit)\b`,
			),
		},
		{
			category: constants.SafetyCategoryIllegalBehavior,
			reason:   "request appears to ask for illegal behavior or evasion",
			patterns: compilePatterns(
				`(?i)\b(how to|teach me|guide me|steps to)\b.{0,80}\b(evade police|launder money|forge|counterfeit|drug trafficking)\b`,
				`(?i)\b(illegal|black market|avoid detection|evade law enforcement)\b`,
			),
		},
		{
			category: constants.SafetyCategorySelfHarm,
			reason:   "request appears to involve self-harm instructions",
			patterns: compilePatterns(
				`(?i)\b(kill myself|suicide|self harm|hurt myself)\b`,
				`(?i)\b(how to|best way to|painless way to)\b.{0,80}\b(die|commit suicide|kill myself)\b`,
			),
		},
		{
			category: constants.SafetyCategorySexualContent,
			reason:   "request appears to ask for sexual content that the assistant should not handle",
			patterns: compilePatterns(
				`(?i)\b(child sexual|sexual content involving minors|minor sexual|underage sexual)\b`,
				`(?i)\b(explicit sexual|pornographic)\b.{0,80}\b(minor|child|underage)\b`,
			),
		},
		{
			category: constants.SafetyCategoryHateHarassment,
			reason:   "request appears to ask for hate, harassment, or abusive targeting",
			patterns: compilePatterns(
				`(?i)\b(write|create|generate)\b.{0,80}\b(hate speech|racial slur|harassment campaign|abusive message)\b`,
				`(?i)\b(target|doxx|harass|threaten)\b.{0,80}\b(person|group|user|employee)\b`,
			),
		},
		{
			category: constants.SafetyCategoryViolence,
			reason:   "request appears to ask for violent instructions",
			patterns: compilePatterns(
				`(?i)\b(how to|build|make|create)\b.{0,80}\b(bomb|explosive|weapon|poison)\b`,
				`(?i)\b(assassinate|kill someone|hurt someone|violent attack)\b`,
			),
		},
		{
			category: constants.SafetyCategoryPrivacy,
			reason:   "request appears to ask for private personal data or privacy-invasive behavior",
			patterns: compilePatterns(
				`(?i)\b(doxx|dox|track someone|stalk|find home address|private phone|private email)\b`,
				`(?i)\b(leak|expose|retrieve|get)\b.{0,80}\b(private data|personal data|home address|phone number)\b`,
			),
		},
	}
}

func compilePatterns(patterns ...string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		compiled = append(compiled, regexp.MustCompile(pattern))
	}
	return compiled
}

func normalize(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func blockedUserMessage(category string) string {
	switch category {
	case constants.SafetyCategorySecrets:
		return "Mình không thể hỗ trợ yêu cầu liên quan đến việc tiết lộ, trích xuất hoặc vượt qua bảo vệ khóa bí mật, token, mật khẩu hay thông tin xác thực."
	case constants.SafetyCategoryCyberAbuse:
		return "Mình không thể hỗ trợ hướng dẫn tấn công, khai thác, đánh cắp thông tin xác thực, malware hoặc phishing."
	case constants.SafetyCategorySelfHarm:
		return "Mình không thể hỗ trợ hướng dẫn tự gây hại. Nếu bạn đang gặp nguy hiểm ngay lúc này, hãy liên hệ người thân hoặc dịch vụ khẩn cấp tại nơi bạn sống."
	default:
		return "Mình không thể hỗ trợ yêu cầu này vì nó thuộc nhóm nội dung nhạy cảm hoặc bị hạn chế."
	}
}
