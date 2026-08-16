# Releases

Dottie release tags use semantic versions such as `v0.1.0`. A pushed tag runs the full test suite and creates native archives for Linux x64/ARM64, macOS Intel/Apple Silicon, and Windows x64. GitHub Releases publishes generated notes and SHA-256 checksums.

To cut a release:

1. Move the relevant `CHANGELOG.md` entries under the new version.
2. Run `make check` on a clean worktree.
3. Commit the release notes.
4. Create and push an annotated tag: `git tag -a v0.1.0 -m "Dottie v0.1.0" && git push origin v0.1.0`.
5. Verify every release-matrix job and the published checksums.

Do not reuse or move a published version tag.

