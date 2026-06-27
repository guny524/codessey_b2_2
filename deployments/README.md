# deployments
- LLM text 응답 api 를 codex 구독을 통해서 사용할 수 있게 해주는 [openai-oauth proxy](https://github.com/EvanZhouDev/openai-oauth) 와
- image 생성을 codex 구독을 통해 api 로 사용할 수 있게 해주는 별도로 만든 [image_proxy](image_proxy/README.md) 를
- docker compose 로 격리화 해서 띄운다
- docker compose 로 서버나 로컬 노트북에 띄우면
  - n8n 을 docker 로 띄우면 로컬에 띄운 api 에 접근시켜서 사용할 수 있겠지만
  - 우리는 Make workflow 로 cloud 에 떠 있는 걸 사용하기 때문에 우리가 따로 서버에 이 docker compose 를 띄워야 한다

## 실행
- docker compose 를 실행하는 호스트 컴퓨터(노트북, 서버) 에서 npm 으로 codex 를 설치하고 `codex login` 이 실행되어서 `~/.codex/auth.json` 위치에 OAuth 토큰이 존재해야 한다
- 그 후 `make up` 실행
- `docker cp` 명령어로 격리된 볼륨에 복사해가기 때문에 워래 OAuth 토큰 파일은 오염되지 않으므로 안심해도 됨
- 사용 종료시 `make down`, `docker compose down -f` 실행되어, 실행 시 만들어졌던 docker volume 도 알아서 정리 된다

## test
`make up` 후 호스트에서 curl 로 확인한다 — **llm_proxy 는 host 8080**, **image_proxy 는 host 8081** (둘 다 컨테이너 내부 포트는 8080)

### llm_proxy (텍스트, OpenAI 호환 chat/completions)
```bash
# 사용 가능 모델 확인
curl -s http://127.0.0.1:8080/v1/models

# 텍스트 응답 (3줄 요약 등에 쓰는 chat/completions)
curl -s http://127.0.0.1:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"gpt-5.4-mini","messages":[{"role":"user","content":"한 줄로 인사해줘"}]}'
```
- 이미지 생성 로컬에서 테스트 [image_proxy 호출 예시](image_proxy/README.md#호출-예시)

## 배포된 서버 주소
- llm_proxy 8080 `210.114.89.138:30000`
- image_proxy 8081 `210.114.89.138:30001`
