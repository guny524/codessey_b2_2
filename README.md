# codyssey_b2_2
- [과제 내용](subject.md)
- [협업 방법](CONTRIBUTE.md)
- [팀 Make Workflow](https://eu1.make.com/2021099/scenarios/6327312/edit)
- [결과 notion](https://app.notion.com/p/Ford-AI-gray-beard-38ca042c32f681fc84eecbe5a71829b7?source=copy_link)

## 팀 작업 요약
- TODO

## 구성 설명
- ![스크린샷](img/screenshot_make.png)
```mermaid
flowchart TD
    sched(["스케줄 트리거 · 매일 자동 실행"]) --> rss["① RSS 뉴스 수집<br/>news.hada.io/rss/news"]
    rss -->|"주제 필터: 제목에 AI·LLM·GPT·생성형 AI·오픈소스·인공지능 포함"| search["② Notion 중복 검색<br/>원문 id로 기존 항목 조회"]
    search -->|"중복 방지: 검색결과 0건(신규)만 통과"| gemini["③ Gemini 요약 (gemini-3.1-flash-lite)<br/>→ JSON: summary · sentiment · image_prompt"]
    gemini --> parse["④ JSON 파싱"]
    parse --> notion["⑤ Notion DB 페이지 생성<br/>title · summary · sentiment · date · 원문 url · id · img"]
    parse -->|"image_prompt"| poll["pollinations.ai<br/>프롬프트 → 썸네일 이미지 URL"]
    poll -->|"img URL"| notion

    search -.->|"실패 시 Break: 재시도 2회 / 15분 간격"| err(("에러 처리"))
    gemini -.-> err
    notion -.-> err
```

## 필터링 기준
### TODO 키워드/태그 목록
- TODO
### 선택 이유
- TODO

## 에러 처리 정책
- TODO

## 부가 요소 설명
- [openai-oauth proxy](https://github.com/EvanZhouDev/openai-oauth) 이 proxy 서버를 띄워 놓으면 LLM text 응답을 codex 구독을 통해서 api endpoint 로 만들 수 있다, 여기선 image 생성 api 는 제공을 안해서 아래 image_proxy 를 별도로 만듬
- [image_proxy](image_proxy/README.md) image 생성을 codex 구독으로 제공하는 api endpoint 를 만들기 위한 도구
- 자세한 내용은 [deployments](deployments/README.md) 배포 사용 방법을 참고
