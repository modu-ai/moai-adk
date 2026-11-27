//! The archive of one tidy item (SPEC-20261126-home-tidy-apply, REQ-3 to REQ-5): the list of
//! its files and folders, the gzip tar that holds them, and the read back.
//!
//! Every path is walked from a handle of the moai folder one name at a time and no link is ever
//! followed. A secret-shaped name is refused by its name, before anything at it is opened.

use super::inventory::Limits;
use super::rules::{self, Class};
use cap_fs_ext::{DirExt, FollowSymlinks, OpenOptionsFollowExt, OpenOptionsMaybeDirExt};
use cap_std::fs::{Dir, DirBuilder, DirBuilderExt, OpenOptions, OpenOptionsExt};
use flate2::{read::GzDecoder, write::GzEncoder, Compression};
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use std::collections::BTreeMap;
use std::fs::{File, Permissions};
use std::io::{self, BufWriter, Read, Write};
use std::os::unix::fs::{MetadataExt, PermissionsExt};
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

/// The list file of one archived item.
pub const LIST: &str = "list.json";
/// The compressed archive of one archived item.
pub const DATA: &str = "data.tar.gz";
const COMPRESSION: &str = "gzip";
const VERSION: u32 = 1;
/// A list larger than this is not read.
const LIST_LIMIT: u64 = 256 * 1024 * 1024;

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum Kind {
    Dir,
    File,
}

/// One file or folder of an item. `path` is relative to the moai folder.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct Row {
    pub path: String,
    #[serde(rename = "type")]
    pub kind: Kind,
    /// File bytes; 0 for a folder.
    pub bytes: u64,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub sha256: Option<String>,
    pub mode: u32,
    pub mtime_secs: i64,
    pub mtime_nanos: u32,
}

/// The content of `list.json`.
#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct List {
    pub version: u32,
    pub item: String,
    pub compression: String,
    pub entries: Vec<Row>,
}

impl Row {
    fn new(path: String, kind: Kind, md: &std::fs::Metadata) -> Self {
        Row {
            path,
            kind,
            bytes: 0,
            sha256: None,
            mode: md.mode() & 0o7777,
            mtime_secs: md.mtime(),
            mtime_nanos: u32::try_from(md.mtime_nsec()).unwrap_or(0),
        }
    }

    /// The modification time the row names.
    pub fn modified(&self) -> SystemTime {
        let nanos = Duration::from_nanos(u64::from(self.mtime_nanos));
        let whole = Duration::from_secs(self.mtime_secs.unsigned_abs());
        if self.mtime_secs >= 0 {
            UNIX_EPOCH + whole + nanos
        } else {
            UNIX_EPOCH - whole + nanos
        }
    }

    /// Whether `now` is still the thing this row listed. A folder's time changes when anything
    /// below it does, so only its type and mode are compared.
    pub fn same_as(&self, now: &Row) -> bool {
        match self.kind {
            Kind::File => self == now,
            Kind::Dir => now.kind == Kind::Dir && self.mode == now.mode,
        }
    }
}

fn why(e: &io::Error) -> String {
    e.kind().to_string()
}

/// Whether a name has the shape of a secret: the same rule the inventory protects by.
pub fn secret(name: &str) -> bool {
    rules::class_of_name(name).0 == Class::Protected
}

/// The names of a relative path. Anything but plain names, or a secret-shaped one, is refused.
pub fn parts(path: &str) -> Result<Vec<&str>, String> {
    let names: Vec<&str> = path.split('/').collect();
    for name in &names {
        if name.is_empty() || *name == "." || *name == ".." || name.contains('\0') {
            return Err(format!("unsafe_path: {path}"));
        }
        if secret(name) {
            return Err(format!("secret_name: {path}"));
        }
    }
    Ok(names)
}

/// The folder that holds the last name of `names`, opened from `root` one name at a time.
pub fn parent(root: &Dir, names: &[&str]) -> Result<Dir, String> {
    let mut dir = root
        .try_clone()
        .map_err(|e| format!("unreadable: {}", why(&e)))?;
    let above = &names[..names.len().saturating_sub(1)];
    for (i, name) in above.iter().enumerate() {
        dir = sub_dir(&dir, name, &names[..=i].join("/"))?;
    }
    Ok(dir)
}

/// The folder `name` of `dir`; a link or anything but a folder is refused.
pub fn sub_dir(dir: &Dir, name: &str, shown: &str) -> Result<Dir, String> {
    let md = dir
        .symlink_metadata(name)
        .map_err(|e| format!("unreadable: {shown}: {}", why(&e)))?;
    if md.file_type().is_symlink() {
        return Err(format!("link: {shown}"));
    }
    if !md.is_dir() {
        return Err(format!("special_file: {shown} is not a folder"));
    }
    dir.open_dir_nofollow(name)
        .map_err(|e| format!("unreadable: {shown}: {}", why(&e)))
}

/// Opens the regular file `name` of `dir` for reading, following no link and never blocking.
pub fn open_file(dir: &Dir, name: &str) -> io::Result<(File, std::fs::Metadata)> {
    let mut options = OpenOptions::new();
    options.read(true).follow(FollowSymlinks::No);
    options.custom_flags(libc::O_NONBLOCK);
    let file = dir.open_with(name, &options)?.into_std();
    let md = file.metadata()?;
    if !md.is_file() {
        return Err(io::Error::new(
            io::ErrorKind::InvalidData,
            "not a regular file",
        ));
    }
    Ok((file, md))
}

/// The folder itself as a file handle, for its own mode, times and sync.
pub fn own_handle(dir: &Dir) -> io::Result<File> {
    let mut options = OpenOptions::new();
    options
        .read(true)
        .follow(FollowSymlinks::No)
        .maybe_dir(true);
    Ok(dir.open_with(".", &options)?.into_std())
}

/// Creates the file `name`, which must not exist, readable by the owner only.
pub fn create_private(dir: &Dir, name: &str) -> io::Result<File> {
    let mut options = OpenOptions::new();
    options
        .write(true)
        .create_new(true)
        .follow(FollowSymlinks::No);
    options.mode(0o600);
    let file = dir.open_with(name, &options)?.into_std();
    file.set_permissions(Permissions::from_mode(0o600))?;
    Ok(file)
}

/// Creates the folder `name`, which must not exist, for the owner only.
pub fn private_dir(dir: &Dir, name: &str) -> io::Result<Dir> {
    let mut builder = DirBuilder::new();
    builder.mode(0o700);
    dir.create_dir_with(name, &builder)?;
    let made = dir.open_dir_nofollow(name)?;
    own_handle(&made)?.set_permissions(Permissions::from_mode(0o700))?;
    Ok(made)
}

/// Reads to the end and gives the digest and the byte count.
pub fn hash(mut from: impl Read) -> io::Result<(String, u64)> {
    copy_hashed(&mut from, &mut io::sink())
}

/// Copies to the end and gives the digest and the byte count of what was copied.
pub fn copy_hashed(from: &mut impl Read, to: &mut impl Write) -> io::Result<(String, u64)> {
    let mut sum = Sha256::new();
    let mut total = 0u64;
    let mut buf = vec![0u8; 128 * 1024];
    loop {
        let n = from.read(&mut buf)?;
        if n == 0 {
            break;
        }
        sum.update(&buf[..n]);
        to.write_all(&buf[..n])?;
        total += n as u64;
    }
    Ok((format!("sha256:{:x}", sum.finalize()), total))
}

struct Walk {
    left: usize,
    deadline: Instant,
    found: Vec<(String, Kind)>,
}

impl Walk {
    /// Looks at names and types only: nothing below `dir` is opened but folders.
    fn visit(&mut self, dir: &Dir, name: &str, path: String) -> Result<(), String> {
        if secret(name) {
            return Err(format!("secret_name: {path}"));
        }
        if self.left == 0 || Instant::now() > self.deadline {
            return Err(format!("over_limit: the walk stopped at {path}"));
        }
        self.left -= 1;
        let md = dir
            .symlink_metadata(name)
            .map_err(|e| format!("unreadable: {path}: {}", why(&e)))?;
        let kind = md.file_type();
        if kind.is_symlink() {
            return Err(format!("link: {path}"));
        }
        if kind.is_file() {
            self.found.push((path, Kind::File));
            return Ok(());
        }
        if !kind.is_dir() {
            return Err(format!("special_file: {path}"));
        }
        let below = sub_dir(dir, name, &path)?;
        self.found.push((path.clone(), Kind::Dir));
        for child in names_of(&below).map_err(|e| format!("unreadable: {path}: {e}"))? {
            self.visit(&below, &child, format!("{path}/{child}"))?;
        }
        Ok(())
    }
}

/// The names in a folder, sorted; a name that is not UTF-8 is an error.
pub fn names_of(dir: &Dir) -> Result<Vec<String>, String> {
    let mut names = Vec::new();
    for entry in dir.entries().map_err(|e| why(&e))? {
        let name = entry.map_err(|e| why(&e))?.file_name();
        names.push(
            name.into_string()
                .map_err(|_| "a name is not UTF-8".to_string())?,
        );
    }
    names.sort();
    Ok(names)
}

/// The rows of the item at `item`, parents first. The whole item is looked at by name and type
/// before any file is opened; one refusal refuses the item.
pub fn list(moai: &Dir, item: &str, limits: Limits) -> Result<Vec<Row>, String> {
    let names = parts(item)?;
    let top = parent(moai, &names)?;
    let mut walk = Walk {
        left: limits.entries.saturating_add(1),
        deadline: Instant::now() + Duration::from_secs(limits.seconds),
        found: Vec::new(),
    };
    let last = names.last().expect("split gives at least one name");
    walk.visit(&top, last, item.to_string())?;
    let mut rows = Vec::with_capacity(walk.found.len());
    for (path, kind) in walk.found {
        rows.push(row(moai, path, kind)?);
    }
    rows.sort_by(|a, b| a.path.cmp(&b.path));
    Ok(rows)
}

fn row(moai: &Dir, path: String, kind: Kind) -> Result<Row, String> {
    let names = parts(&path)?;
    let dir = parent(moai, &names)?;
    let name = *names.last().expect("split gives at least one name");
    let unreadable = |e: io::Error| format!("unreadable: {path}: {}", why(&e));
    match kind {
        Kind::Dir => {
            let own = sub_dir(&dir, name, &path)?;
            let md = own_handle(&own)
                .and_then(|f| f.metadata())
                .map_err(unreadable)?;
            Ok(Row::new(path, kind, &md))
        }
        Kind::File => {
            let (file, md) = open_file(&dir, name).map_err(unreadable)?;
            let (sum, bytes) = hash(file).map_err(unreadable)?;
            let mut row = Row::new(path, kind, &md);
            row.sha256 = Some(sum);
            row.bytes = bytes;
            Ok(row)
        }
    }
}

/// Writes the list and the compressed archive of `rows` into `dir`, both new files.
pub fn write(moai: &Dir, dir: &Dir, item: &str, rows: &[Row]) -> Result<(), String> {
    let fail = |e: io::Error| format!("archive_not_written: {}", why(&e));
    let data = create_private(dir, DATA).map_err(fail)?;
    let gz = GzEncoder::new(BufWriter::new(data), Compression::default());
    let mut tar = tar::Builder::new(gz);
    for row in rows {
        let mut head = tar::Header::new_gnu();
        head.set_mode(row.mode);
        head.set_mtime(u64::try_from(row.mtime_secs).unwrap_or(0));
        match row.kind {
            Kind::Dir => {
                head.set_entry_type(tar::EntryType::Directory);
                head.set_size(0);
                tar.append_data(&mut head, &row.path, io::empty())
                    .map_err(fail)?;
            }
            Kind::File => {
                head.set_entry_type(tar::EntryType::Regular);
                head.set_size(row.bytes);
                let names = parts(&row.path)?;
                let from = parent(moai, &names)?;
                let (file, _) = open_file(&from, names[names.len() - 1]).map_err(fail)?;
                let mut counted = Counted {
                    from: file.take(row.bytes),
                    sum: Sha256::new(),
                    total: 0,
                };
                tar.append_data(&mut head, &row.path, &mut counted)
                    .map_err(fail)?;
                let sum = format!("sha256:{:x}", counted.sum.finalize());
                if counted.total != row.bytes || Some(sum) != row.sha256 {
                    return Err(format!("source_changed: {}", row.path));
                }
            }
        }
    }
    let data = tar
        .into_inner()
        .and_then(GzEncoder::finish)
        .and_then(|w| w.into_inner().map_err(|e| e.into_error()))
        .map_err(fail)?;
    data.sync_all().map_err(fail)?;
    let list = List {
        version: VERSION,
        item: item.to_string(),
        compression: COMPRESSION.to_string(),
        entries: rows.to_vec(),
    };
    let text = serde_json::to_vec_pretty(&list).map_err(|e| format!("archive_not_written: {e}"))?;
    let mut file = create_private(dir, LIST).map_err(fail)?;
    file.write_all(&text).map_err(fail)?;
    file.sync_all().map_err(fail)?;
    own_handle(dir).and_then(|f| f.sync_all()).map_err(fail)
}

struct Counted<R> {
    from: R,
    sum: Sha256,
    total: u64,
}

impl<R: Read> Read for Counted<R> {
    fn read(&mut self, buf: &mut [u8]) -> io::Result<usize> {
        let n = self.from.read(buf)?;
        self.sum.update(&buf[..n]);
        self.total += n as u64;
        Ok(n)
    }
}

/// Reads `list.json` of the archived item in `dir`.
pub fn read_list(dir: &Dir) -> Result<List, String> {
    let (file, md) =
        open_file(dir, LIST).map_err(|e| format!("the list is not readable: {}", why(&e)))?;
    if md.len() > LIST_LIMIT {
        return Err("the list is over its size limit".into());
    }
    let mut text = Vec::new();
    file.take(LIST_LIMIT)
        .read_to_end(&mut text)
        .map_err(|e| format!("the list is not readable: {}", why(&e)))?;
    let list: List =
        serde_json::from_slice(&text).map_err(|e| format!("the list does not parse: {e}"))?;
    if list.version != VERSION || list.compression != COMPRESSION {
        return Err("the list has a version or compression this build does not read".into());
    }
    Ok(list)
}

/// Whether a list names only the item and what lies below it, once each, parents included.
pub fn check_rows(list: &List) -> Result<(), String> {
    if parts(&list.item)?.first() == Some(&"archive") {
        return Err("the list names a place inside the archive folder".into());
    }
    let mut seen: BTreeMap<&str, Kind> = BTreeMap::new();
    let mut rows: Vec<&Row> = list.entries.iter().collect();
    rows.sort_by(|a, b| a.path.cmp(&b.path));
    let below = format!("{}/", list.item);
    for row in rows {
        parts(&row.path)?;
        if row.path != list.item {
            if !row.path.starts_with(&below) {
                return Err(format!("`{}` leaves the item", row.path));
            }
            let above = row.path.rsplit_once('/').map_or("", |(a, _)| a);
            if seen.get(above) != Some(&Kind::Dir) {
                return Err(format!("`{}` has no listed folder above it", row.path));
            }
        }
        if row.mode > 0o7777 || row.mtime_nanos >= 1_000_000_000 {
            return Err(format!("`{}` has a mode or time out of range", row.path));
        }
        if (row.kind == Kind::File) != row.sha256.is_some() {
            return Err(format!("`{}` lacks its hash", row.path));
        }
        if seen.insert(&row.path, row.kind).is_some() {
            return Err(format!("`{}` is listed twice", row.path));
        }
    }
    if !seen.contains_key(list.item.as_str()) {
        return Err("the list does not name the item itself".into());
    }
    Ok(())
}

/// One entry of an archive: its row and its content.
type Visit<'a> = dyn FnMut(&Row, &mut dyn Read) -> Result<(String, u64), String> + 'a;

/// Goes through the whole archive in `dir`: each entry must be one row of the same type, size
/// and hash, every row must be there once, and the compressed stream must be sound to its end.
/// `visit` reads the content of a file entry and says what it read.
pub fn each_entry(dir: &Dir, rows: &[Row], visit: &mut Visit<'_>) -> Result<(), String> {
    let bad = |e: io::Error| format!("the archive does not read: {}", why(&e));
    let (file, _) = open_file(dir, DATA).map_err(bad)?;
    let mut left: BTreeMap<&str, &Row> = rows.iter().map(|r| (r.path.as_str(), r)).collect();
    let mut tar = tar::Archive::new(GzDecoder::new(file));
    for entry in tar.entries().map_err(bad)? {
        let mut entry = entry.map_err(bad)?;
        let path = entry.path().map_err(bad)?.to_string_lossy().into_owned();
        let path = path.trim_end_matches('/');
        let Some(row) = left.remove(path) else {
            return Err(format!("the archive holds `{path}`, which the list lacks"));
        };
        let kind = entry.header().entry_type();
        match row.kind {
            Kind::Dir if kind.is_dir() => {}
            Kind::File if kind.is_file() => {
                let (sum, bytes) = visit(row, &mut entry)?;
                if bytes != row.bytes || Some(sum) != row.sha256 {
                    return Err(format!("`{path}` differs from the list"));
                }
            }
            _ => return Err(format!("`{path}` has another type than the list says")),
        }
    }
    if let Some(path) = left.keys().next() {
        return Err(format!("the archive lacks `{path}`"));
    }
    // The checksum of the compressed stream stands after the last entry.
    io::copy(&mut tar.into_inner(), &mut io::sink()).map_err(bad)?;
    Ok(())
}

/// Reads the whole archive back against `rows` without writing anything.
pub fn check(dir: &Dir, rows: &[Row]) -> Result<(), String> {
    each_entry(dir, rows, &mut |_, content| {
        hash(content).map_err(|e| format!("the archive does not read: {}", why(&e)))
    })
}
