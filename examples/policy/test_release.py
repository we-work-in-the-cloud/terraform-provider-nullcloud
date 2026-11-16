"""Verify a full GoReleaser snapshot: python3 examples/policy/test_release.py [dist]."""

import hashlib
import json
from pathlib import Path
import sys
import zipfile


dist = Path(sys.argv[1] if len(sys.argv) > 1 else "dist")
version = json.loads((dist / "metadata.json").read_text())["version"]
# Keep the released provider matrix intact, including GoReleaser's unsupported-pair filtering.
platforms = {
    "darwin_amd64", "darwin_arm64", "freebsd_amd64",
    "linux_386", "linux_amd64", "linux_arm", "linux_arm64", "linux_s390x",
    "windows_386", "windows_amd64", "windows_arm64",
}
expected = {
    f"{name}_{version}_{platform}.zip"
    for name in ("terraform-provider-nullcloud", "nullcloud-policy")
    for platform in platforms
}
actual = {path.name for path in dist.glob("*.zip")}
assert actual == expected, f"archive matrix: missing={expected - actual}, extra={actual - expected}"
manifest = dist / f"terraform-provider-nullcloud_{version}_SHA256SUMS"
checksums = dict(line.split()[::-1] for line in manifest.read_text().splitlines())
assert set(checksums) == expected, "checksum manifest must cover exactly the release archives"
for name in sorted(expected):
    archive = dist / name
    assert hashlib.sha256(archive.read_bytes()).hexdigest() == checksums[name], name
    policy = name.startswith("nullcloud-policy_")
    executable = "nullcloud-policy" if policy else f"terraform-provider-nullcloud_v{version}"
    if "_windows_" in name:
        executable += ".exe"
    with zipfile.ZipFile(archive) as contents:
        members = set(contents.namelist())
        assert members == {"LICENSE", "README.md", "changelog.md", executable}, (name, members)
        assert contents.getinfo(executable).external_attr >> 16 & 0o111, (name, "not executable")
    print(f"OK {name}: checksum, executable, archive separation")
print(f"Verified {len(expected)} archives across {len(platforms)} platforms; manifest {manifest.name}")
