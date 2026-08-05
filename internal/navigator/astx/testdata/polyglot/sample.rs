struct Point { x: i32 }
trait Named { fn name(&self) -> &str; }
fn origin() -> Point { Point { x: 0 } }
