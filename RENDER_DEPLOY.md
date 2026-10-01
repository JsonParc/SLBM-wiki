# Render 배포

이 저장소는 Go 기반 SLBM 위키 엔진입니다. Render에서는 `render.yaml`을 사용해 빌드와 실행을 설정합니다.

## 배포 순서

1. GitHub에서 `JsonParc/SLBM-wiki` 저장소를 엽니다.
2. Render 대시보드에서 **New + > Blueprint**를 선택합니다.
3. GitHub 저장소 `JsonParc/SLBM-wiki`를 연결합니다.
4. 저장소의 `render.yaml`을 확인하고 배포를 시작합니다.
5. 첫 실행이 끝나면 Render가 발급한 `onrender.com` 주소를 엽니다.

`render.yaml`은 Go 빌드, Render가 제공하는 `$PORT`로 서버 실행, `/` 헬스 체크를 처리합니다.

## 데이터 보존

영속 디스크가 없으면 재배포·재시작 때마다 사이트가 저장소에 커밋된 `data.db`와 `seed/images` 상태로 돌아갑니다. 사이트에서 가입한 계정, 편집, 직책, 클랜 벌점이 모두 사라집니다.

### 디스크 없이 운영할 때

재배포 전에 관리자 계정으로 **사이트 백업·복원**(`/clan/backup`)에서 백업 zip을 받고, 배포가 끝나면 같은 화면에서 그 zip을 불러옵니다. 문서, 편집 기록, 계정, 직책, 벌점, 이미지가 모두 돌아옵니다.

### 영속 디스크로 운영할 때

유료 인스턴스에서 디스크를 붙이면 재배포 후에도 데이터가 그대로 남습니다.

1. Render 서비스의 **Settings > Instance Type**을 Starter 이상으로 바꿉니다.
2. **Disks**에서 디스크를 추가하고 Mount Path를 `/opt/render/project/src/data`로 지정합니다. 1GB면 충분합니다.
3. **Environment**에 `NAMU_DB` = `data/data`를 추가합니다.
4. 기존 사이트 데이터를 옮기려면 전환 직전에 백업 zip을 받아 두었다가, 전환 후 첫 배포가 끝나면 불러옵니다.

디스크를 처음 붙이면 서버가 시작할 때 커밋된 `data.db`와 `seed/images`를 디스크로 한 번 복사합니다. 디스크에 이미 있는 파일은 덮어쓰지 않습니다.

## 주의사항

`data.db`에는 비밀번호 해시와 로그인 키가 들어 있습니다. 저장소를 비공개로 유지하세요. 백업 zip도 같은 정보를 담고 있으니 공유하지 마세요.

## 업데이트

소스를 수정한 뒤 GitHub의 기본 브랜치에 push하면 Render가 자동 재배포합니다. 배포 로그에서 `Build successful`과 서버의 `Run in http://0.0.0.0:<PORT>` 로그를 확인하세요.
