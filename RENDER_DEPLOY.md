# Render 배포

이 저장소는 Go 기반 SLBM 위키 엔진입니다. Render에서는 `render.yaml`을 사용해 빌드와 실행을 설정합니다.

## 배포 순서

1. GitHub에서 `JsonParc/SLBM-wiki` 저장소를 엽니다.
2. Render 대시보드에서 **New + > Blueprint**를 선택합니다.
3. GitHub 저장소 `JsonParc/SLBM-wiki`를 연결합니다.
4. 저장소의 `render.yaml`을 확인하고 배포를 시작합니다.
5. 첫 실행이 끝나면 Render가 발급한 `onrender.com` 주소를 엽니다.

`render.yaml`은 다음을 자동으로 처리합니다.

- Go 의존성 설치 및 엔진 빌드
- Render가 제공하는 `$PORT`로 서버 실행
- `/` 헬스 체크
- SQLite 데이터와 검색 인덱스가 재배포 후에도 남도록 `/data` 영속 디스크 사용

## 주의사항

영속 디스크는 Render 무료 웹 서비스에서 제공되지 않을 수 있습니다. 무료 플랜으로 실행하면 재시작·재배포 때 `data/` 아래 SQLite 데이터와 검색 인덱스가 초기화될 수 있습니다. 회원 계정과 문서를 보존하려면 영속 디스크가 포함된 플랜을 사용하세요.

첫 배포 후 관리자 계정과 사이트 설정을 확인하고, 기존 로컬 `data.db`를 옮겨야 한다면 서버를 중지한 뒤 백업본을 기준으로 별도 마이그레이션을 진행하세요. 운영 중인 데이터베이스를 Git에 커밋하거나 공개 저장소에 올리면 안 됩니다.

## 업데이트

소스를 수정한 뒤 GitHub의 기본 브랜치에 push하면 Render가 자동 재배포합니다. 배포 로그에서 `Build successful`과 서버의 `Run in http://0.0.0.0:<PORT>` 로그를 확인하세요.
