# LLM, Vision model API 어떻게 할지 고민/테스트
- 이슈 주소: `https://github.com/guny524/codessey_b2_2/issues/1`
- 과제(Codessey b2_2 "뉴스 요약 자동화 워크플로우")가 호출할 LLM 텍스트 요약 + 이미지 생성 API를, 비용/제한 없이 OpenAI 호환 엔드포인트로 제공하는 방법 결정 및 구현

## 1. 배경(현재 이슈의 대략적인 이전 맥락)
- 과제는 RSS 수집 -> 주제 필터 -> AI 3줄 요약 -> 노션 저장의 매일 무인 자동화(n8n/Make). 보너스로 썸네일 이미지 생성
- 제약: 생성형 AI는 "OpenAI API 또는 이에 준하는 모델"(OpenAI 호환 endpoint), 자동화 툴은 Make/Zapier/n8n 자유
- 조민기 담당 = "AI 요약 프롬프트 / API". 워크플로우가 호출할 AI 엔드포인트 제공이 본 작업 범위
#### 1-1. 참고해야 할 제약/의도
- 무료 API(OpenRouter/Gemini 등)는 개발/테스트 반복에서 한도가 빠르게 소진됨 -> 본인 ChatGPT 구독(Codex OAuth)을 OpenAI 호환 endpoint로 노출해 무과금으로 쓰는 방향
- 평가 기간 동안만 엔드포인트를 켜두면 되고, endpoint 주소는 환경변수로 빼서 교체 가능하게 둔다

---

## 2. 요구사항(구현하고자 하는 필요한 기능)
### 2-1. 텍스트 요약 엔드포인트
- ChatGPT 구독 기반 `/v1/chat/completions` (OpenAI 호환), 호출당 과금 없음
### 2-2. 이미지 생성 엔드포인트
- `/v1/images/generations`로 요청하면 response로 이미지(base64) 반환, ChatGPT 구독(codex 내장 image_gen, gpt-image) 기반 무과금
- 공식 OpenAI images API 스펙 준수: `prompt`, `size`(WxH) 지원
#### 2-2-1. 이 기능을 구현하며 신경 써야 하는 부분
- codex는 범용 에이전트라 bypass 모드에서 shell 권한 보유 -> 프롬프트를 OS shell이 아닌 인자로 분리 전달, 고정 템플릿으로 감싸 주입 표면 최소화
- 무인 자동화를 위해 승인 개입 없이(`--dangerously-bypass-approvals-and-sandbox`) 동작해야 함
#### 2-2-2. 우선적으로 참조할 파일
- `image_proxy/` (신규 Go 서버)
- 참조 컨벤션: `/Users/min-jo/go/src/gitlab.com/01ai/eng/aiauto/aiauto/back` (Makefile/template.mk, .golangci.yml, internal 레이아웃, main_test.go)

---

# AI 결과

## 3. (AI가 확인한) 기존 코드/구현의 핵심내용들/의도들
- 신규 컴포넌트, 기존 코드 없음 (repo는 README, CONTRIBUTE만 존재)
- 환경 실측: go 1.26.1, node v22.18.0, npx 11.11.0, codex-cli 0.139.0, make 3.81, golangci-lint 2.11.3
- codex 인증 실측: `~/.codex/auth.json`에 OAuth 토큰(id/access/refresh/account_id) 존재, `OPENAI_API_KEY` 미설정 -> 구독(OAuth) 인증 확인
- 참조 프로젝트 컨벤션: 각 서비스가 `Makefile`(루트 `template.mk` 변수 + lint/test/build 타깃), `.golangci.yml`(v2, strict), `internal/<pkg>/`에 코드+`_test.go` 동거, `main_test.go`로 main 헬퍼까지 테스트

---

## 4. 생각한 수정 방안들
### 4-1. 통합 멀티모달 1모델(Janus-Pro/Emu3/Chameleon)을 vLLM으로 온프레미스 서빙
- 장점: 모델 1개로 텍스트+이미지, 운영 단순
- 단점: A4000(16GB)에 올릴 통합 모델은 구형 base(Janus-Pro=DeepSeek-LLM-7B-base, MMLU~48)라 텍스트 요약 품질 불리, 이미지도 384px라 썸네일 부적합
- 판정: 기각 (채점 핵심인 텍스트 품질과 썸네일 품질 양축 모두 열세)
### 4-2. 텍스트=oauth proxy, 이미지=A4000 diffusion(SDXL/FLUX, ComfyUI) 분리 서빙
- 장점: ToS 안전, 로컬 완전 제어
- 단점: GPU 모델 운영 부담, gpt-image보다 품질 낮음
- 판정: 보류(폴백). ToS 회피가 목적일 때만 의미
### 4-3. 표준 `npx openai-oauth` 프록시로 이미지까지 처리
- 장점: 별도 코드 불요
- 단점: 실측 결과 프록시는 `/v1/responses`, `/v1/chat/completions`, `/v1/models`만 노출, `/v1/images/generations` 없음
- 판정: 기각 (이미지 엔드포인트 부재). 단 텍스트 경로로는 채택
### 4-4. codex 내장 image_gen을 `/v1/images/generations` 뒤에 래핑하는 자작 Go 서버
- 장점: gpt-image 품질, 구독 무과금, OpenAI 호환이라 n8n/Make에서 그대로 호출, Make 클라우드도 포트포워딩으로 가능
- 단점: codex headless 실행 + 결과 파일 회수 로직 필요, 구독 usage 소모(이미지 3-5배)
- 판정: 채택. `codex exec ... -C <tmp>`로 승인 개입 없이 1254px PNG 생성 + 경로 회수 실측 검증됨

---

## 5. 최종 결정된 수정 방안 (사용자 승인: "둘 다 진행")
- 텍스트는 기존 `npx openai-oauth` 프록시(4-3), 이미지는 자작 Go 서버(4-4). 둘 다 ChatGPT 구독(OAuth) 무과금, OpenAI 호환
### 5-1. 텍스트 경로 = openai-oauth 프록시 재사용
- `deployments` docker compose 의 llm_proxy 서비스가 `npx -y openai-oauth` 실행 (호스트 8080 -> 컨테이너 8080)
- 이유: 이미 존재하는 도구로 `/v1/chat/completions`가 구독으로 동작(모델 gpt-5.5/5.4/5.4-mini 실측), 직접 구현 불필요
### 5-2. 이미지 경로 = image_proxy
- `image_proxy/internal/server`(Gin)가 `/v1/images/generations` 수신 -> `internal/codex`의 `CLI.Generate(prompt, size)`가 `codex exec --dangerously-bypass-approvals-and-sandbox -C <tmp> "<%q 템플릿 + size 힌트>"` 호출 -> PNG 회수(`out.png` 우선, 없으면 `~/.codex/generated_images` 최신) -> base64로 OpenAI 형태 `{created, data:[{b64_json}]}` 응답
- 이유: 공식 OpenAI images API 형태(b64_json) 유지하면서 구독 기반 gpt-image 품질 확보, 프롬프트 인자 분리 + 템플릿으로 주입 표면 축소. size는 codex에 best-effort 힌트로만 전달(Go 리사이즈/보간 없음 — n8n/Notion이 정확 크기를 요구하지 않으므로 imageutil 제거)
### 5-3. 배포/연결
- `deployments/*/run.sh` + `deployments/run-all.sh`로 두 서버 기동. self-host n8n은 localhost, Make 클라우드/원격은 포트포워딩 + `IMAGE_PROXY_API_KEY` 보호

---

## 6. 코드 수정 요약
- Go 서버 신규 구현 + 참조 컨벤션 정합 + 배포 스크립트/문서
### 6-1. Go 이미지 프록시 (`image_proxy/`)
- [x] `go.mod`/`main.go`/`main_test.go` 모듈 + 설정(env) + 기동, `envOrDefault`/`newHTTPServer`/`newGenerator` 헬퍼 테스트 / 검증: `make test`
- [x] `internal/codex/codex.go`(+`_test.go`) codex CLI 이미지 생성기(Generator 인터페이스, 인자 분리 호출, out.png/generated_images 회수) / 검증: codex_test
- [x] `internal/server/server.go`(+`_test.go`) **Gin** 기반 `/v1/images/generations`·`/healthz`, size 형식검증(WxH), 선택적 bearer 인증, OpenAI 형태 응답 / 검증: server_test
- [x] HTTP 프레임워크로 `gin-gonic/gin` 추가, `codex.Generate(prompt, size)`로 size 힌트 전달, **imageutil/리사이즈 제거** / 검증: `make test`
- [x] `Makefile`(template.mk 병합형) + `.golangci.yml`(참조 복사, prefix 교체) / 검증: `make test` 시 golangci-lint 0 issues
### 6-2. 배포/문서
- [x] `deployments/llm_proxy/run.sh`, `deployments/image_proxy/run.sh`, `deployments/run-all.sh` / 검증: `bash -n` + chmod +x
- [x] `image_proxy/README.md`(코드) + `deployments/README.md`·`deployments/llm_proxy/README.md`·`deployments/image_proxy/README.md`(배포/연결) 작성 / 검증: 수동 리뷰
### 6-3. 검증(실측)
- [x] `make test`: go fmt + golangci-lint 0 issues + `go test -race` 4패키지 전부 ok
- [x] e2e: 서버(Gin) 기동 -> `/healthz` ok, `/v1/images/generations` -> OpenAI 형태 JSON(`{created,data:[{b64_json}]}`) + 유효 PNG 확인. size는 힌트라 `768x768` 요청에 1254px 반환(리사이즈 없음, 의도된 동작)
- [ ] (후속) n8n 워크플로우에서 두 엔드포인트 실제 호출 end-to-end / 검증: n8n 실행 + 노션 저장 확인 (워크플로우 담당 몫)

---

## 7. 문제 해결에 참고
- issue: https://github.com/guny524/codessey_b2_2/issues/1
- codex CLI image gen: https://developers.openai.com/codex/cli/features , https://developers.openai.com/codex/auth
- openai-oauth 프록시(텍스트 전용): https://github.com/EvanZhouDev/openai-oauth , https://news.hada.io/topic?id=28569
- OAuth 이미지 제약 논의: https://github.com/openclaw/openclaw/issues/71179
- 구독 기반 이미지 CLI 사례: https://github.com/leeguooooo/chatgpt-imagegen

## 8. 수정사항 요약 (done_ 시 작성)
- (모든 checkbox 완료 + n8n 연동 확인 후 `done_` prefix 부여 시점에 작성)
