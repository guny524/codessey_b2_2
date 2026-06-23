# Contributing
팀 협업 위한 정리 문서

## 팀/개인 역할
- 사실 작업 영역을 명확하게 나눌 수는 없고
  - 그 때 그 때 필요한 부분은 github 에 이슈로 만들어서 진행하자
### Workflow 자동화
- [김재은](wodms0325@gmail.com)
- [박정욱](wjddnr1234@cau.ac.kr)
- [박찬웅](woong6041@gmail.com)
### AI 요약 프롬프트 / API
- [조민기](guny524@gmail.com)
- [이태규](steelreto@gmail.com)

## git flow
- git flow 라는 협업할 때 branch 를 나눠서 작업하는 방법론이 있는데 아래 참고해보시면 좋을 것 같습니다
  - [Vincent Driessen 의 git flow](https://nvie.com/posts/a-successful-git-branching-model/)
  - [Atlassian git flow](https://www.atlassian.com/git/tutorials/comparing-workflows/gitflow-workflow)
- 요즘은 ai 많이 사용하니까 'git worktree with ai' 같은 키워드로 검색해서 어떤 식으로 사용하는지 꼭 찾아보길 추천합니다
  - 로컬에서 ai 를 병렬로 돌릴 때 디렉토리/branch 를 어떻게 충돌 안 나게 관리 하냐에 대한 내용입니다

### branch 는 main 과 개별 이슈 branch 만
- 원래는 개발과/배포가 별도 구분이 필요할 때, develop/release 라는 이름의 branch 를 유지하는 게 필요한데
  - 저희는 배포는 따로 없으니까, 그냥 main 에서 각자 이슈 만든 후, 이슈 안에서 branch 만들어서 1인당 1이슈당 branch 하나씩 만들어서 작업해주시면 될 것 같습니다
### merge squash 관련
- MR merge 시에는 squash 를 사용하지 않고, 원본 git log 가 전부 main 에 들어가게 merge 해주시면 됩니다
  - 이번 프로젝트는 내용이 그렇게 많이 않기 때문에 굳이 squash 는 안해도 될 것 같습니다
- MR merge 후에는 issue 에 기존 branch 를 삭제할거냐 버튼이 뜰 텐데, 종료된 이슈에 대한 branch 는 삭제해주시면 됩니다
  - 어차피 merge 시에 squash 하지 않고 main 에 merge 하면 main 에서 git log 했을 때 내역이 다 뜨기 때문에, 완료된 개별 이슈에 대한 branch 는 삭제해도 됩니다 (squash 를 한다면, 삭제를 안 하던가 다른 전략을 고려해야함)
### review 관련
- MR 에서 최소 2명 한테 review 받고 merge 하기

## 참고
- 구조 설명이 필요할 시 [mermaid](https://mermaid.ai/open-source/syntax/flowchart.html)(README 문서용) 나 [marp](https://www.npmjs.com/package/@marp-team/marp-cli)(발표 ppt, pdf 용) 를 참고해서 사용해보자
  - Workflow 자동화 도구 캡쳐해서 써도 됨
