# `package-inspect`

operation `package.inspect` · 코드 `src/ops/package_inspect.rs`

`package-inspect PACKAGE` 는 패키지 폴더의 tree digest(`moai-package-tree-v1`)를 계산한다. 파일을 읽기만 하고 쓰지 않으며 stdin 도 읽지 않는다. 결과: `digest`, `digest_algorithm`, `publisher_trust`(`not_verified`), `host_loaded`(`false`).
