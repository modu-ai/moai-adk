//! The user-only CLI defaults file `<HOME state folder>/config/launcher.yaml` (#2871 follow-up,
//! S1 and S3). One reader for everything that looks at it: the launcher (`session`) and
//! `moai setting` (`setting`). Like the project profile's `.yaml` file it accepts only the JSON
//! subset, parsed by serde_json (not general YAML). Reads only; errors never quote the file's
//! content.

use serde_json::{Map, Value};
use std::fs;
use std::path::{Path, PathBuf};

/// The file, relative to the HOME state folder.
pub const FILE: &str = "config/launcher.yaml";

/// Where the file is for `home`: under the folder [`crate::home_root::path`] names.
pub fn path(home: &Path) -> PathBuf {
    crate::home_root::path(home).join(FILE)
}

/// The root object of the file; `None` when the file is absent. The entry is checked before it
/// is read: directories, symlinks and special files (including FIFOs, which could block a read)
/// are not configuration files.
pub fn read_root(home: &Path) -> Result<Option<Map<String, Value>>, String> {
    let path = path(home);
    match fs::symlink_metadata(&path) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(None),
        Err(_) => return Err("user launcher configuration cannot be read".into()),
        Ok(metadata) if !metadata.file_type().is_file() => {
            return Err("user launcher configuration must be a regular file".into())
        }
        Ok(_) => {}
    }
    let bytes = fs::read(&path).map_err(|_| "user launcher configuration cannot be read")?;
    let value: Value = serde_json::from_slice(&bytes)
        .map_err(|_| "user launcher configuration must use valid JSON-subset syntax")?;
    match value {
        Value::Object(root) => Ok(Some(root)),
        _ => Err("user launcher configuration must be an object".into()),
    }
}
