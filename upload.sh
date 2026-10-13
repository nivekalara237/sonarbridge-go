#!/usr/bin/env bash
# Uploads every release file listed in a goreleaser artifacts.json to the local registry,
# then publishes the version. Run it after `goreleaser release`.
#
# Usage: upload.sh [--latest] [--dry-run] [--root DIR] [--version V] <vcs> [artifacts.json]
#   --latest     also make this version the vcs's latest (publish?latest=true)
#   --dry-run    print what would be sent, send nothing (no token needed)
#   --root DIR   directory the "path" fields are relative to (goreleaser's working directory);
#                guessed from where artifacts.json sits, else the current directory
#   --version V  version to publish; default: "tag" (else "version") of metadata.json
#   artifacts.json defaults to dist/artifacts.json
#
#   LOCAL_REGISTRY_URL           e.g. http://localhost:8077
#   LOCAL_REGISTRY_UPLOAD_TOKEN  the server's token; sent on stdin, never on a command line
#
# Each file is sent under its artifacts.json "name" (its "path" is only where it is read
# from), with X-Checksum-Sha256 when the entry carries a sha256 checksum. Build outputs
# (type Binary without extra.Format) are skipped: they are the raw files the archives were
# made from, are not in the checksums file, and several share one name. artifacts.json
# itself is sent last. Needs bash, curl and jq.
set -euo pipefail

die() { echo "upload.sh: $*" >&2; exit 1; }
usage() { sed -n '4,11s/^# \{0,1\}//p' "$0" >&2; exit 2; }

latest=false dry=false root= version=
while [ $# -gt 0 ]; do
	case $1 in
	--latest) latest=true ;;
	--dry-run) dry=true ;;
	--root) [ $# -ge 2 ] || usage; root=$2; shift ;;
	--root=*) root=${1#*=} ;;
	--version) [ $# -ge 2 ] || usage; version=$2; shift ;;
	--version=*) version=${1#*=} ;;
	-h | --help) usage ;;
	--) shift; break ;;
	-*) echo "upload.sh: unknown option $1" >&2; usage ;;
	*) break ;;
	esac
	shift
done
[ $# -ge 1 ] && [ $# -le 2 ] || usage
vcs=$1
artifacts=${2:-dist/artifacts.json}

command -v jq >/dev/null || die "jq is required"
command -v curl >/dev/null || die "curl is required"
[ -f "$artifacts" ] || die "$artifacts not found: run goreleaser first"
jq -e 'type == "array"' "$artifacts" >/dev/null 2>&1 || die "$artifacts is not a JSON array"
if ! $dry; then
	url=${LOCAL_REGISTRY_URL:?LOCAL_REGISTRY_URL is not set}
	url=${url%/}
	: "${LOCAL_REGISTRY_UPLOAD_TOKEN:?LOCAL_REGISTRY_UPLOAD_TOKEN is not set}"
fi

# Root: the "path" fields are relative to goreleaser's working directory. The metadata (or
# checksums) entry sits in the dist directory, beside artifacts.json: strip its relative dir.
dist=$(cd "$(dirname "$artifacts")" && pwd)
if [ -z "$root" ]; then
	rel=$(jq -r 'first(.[] | select(.type == "Metadata" or .type == "Checksum") | .path) // empty' "$artifacts")
	rel=$(dirname "${rel:-.}")
	case $rel in
	/*) root=/ ;;
	.) root=$dist ;;
	*) case $dist in */"$rel") root=${dist%/"$rel"} ;; *) root=$PWD ;; esac ;;
	esac
fi
[ -d "$root" ] || die "--root $root is not a directory"

# resolve <path>: the file on disk for an artifacts.json path.
resolve() {
	case $1 in
	/*) printf '%s\n' "$1" ;;
	*) printf '%s\n' "$root/$1" ;;
	esac
}

# The release files: name, path, sha256 (may be empty), tab-separated.
list=$(jq -r '
	.[]
	| select(
		(.type == "Binary" and (.extra.Format // "") != "")
		or (.type as $t | ["Archive", "File", "Source", "Linux Package", "Checksum",
			"Signature", "Certificate", "SBOM", "Metadata"] | index($t))
	)
	| [.name, .path,
	   ((.extra.Checksum // "") | if startswith("sha256:") then .[7:] else "" end)]
	| @tsv' "$artifacts")
[ -n "$list" ] || die "no uploadable file in $artifacts"

dups=$(printf '%s\n' "$list" | cut -f1 | sort | uniq -d)
[ -z "$dups" ] || die "several entries of $artifacts share a name: $(echo $dups)"
printf '%s\n' "$list" | cut -f1 | grep -qx artifacts.json && die "$artifacts lists a file named artifacts.json"

# Check every file before sending anything, so a missing one does not leave a half upload.
while IFS=$'\t' read -r name path sum; do
	[ -f "$(resolve "$path")" ] || die "$name: $(resolve "$path") not found (wrong --root?)"
done <<<"$list"
printf '%s\n' "$list" | cut -f1 | grep -qi 'checksums\.txt$' || die "$artifacts lists no checksums file"

if [ -z "$version" ]; then
	meta=$(printf '%s\n' "$list" | awk -F'\t' '$1 == "metadata.json" { print $2; exit }')
	[ -n "$meta" ] || die "$artifacts lists no metadata.json: pass --version"
	version=$(jq -r '.tag // .version // empty' "$(resolve "$meta")")
	[ -n "$version" ] || die "no tag nor version in $(resolve "$meta"): pass --version"
fi

# call <curl args...>: sends the token on stdin, prints the server's error body on failure.
call() {
	if ! resp=$(printf 'Authorization: Bearer %s\n' "$LOCAL_REGISTRY_UPLOAD_TOKEN" |
		curl --fail-with-body -sS -H @- "$@"); then
		echo "$resp" >&2
		exit 1
	fi
}

# put <name> <file> [sha256]
put() {
	local enc
	enc=$(jq -rn --arg n "$1" '$n | @uri')
	echo "PUT $1 <- $2${3:+ (sha256 $3)}"
	$dry && return
	call ${3:+-H "X-Checksum-Sha256: $3"} -T "$2" "$url/$vcs/artifacts/$version/files/$enc"
}

echo "== $vcs $version from $artifacts (root $root)"
while IFS=$'\t' read -r name path sum; do
	put "$name" "$(resolve "$path")" "$sum"
done <<<"$list"
put artifacts.json "$artifacts"

echo "POST publish $vcs/$version (latest=$latest)"
$dry && exit 0
call -X POST "$url/$vcs/artifacts/$version/publish?latest=$latest"
echo "$resp"
