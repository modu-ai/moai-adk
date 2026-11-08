import pathlib, sys, unittest
sys.path.insert(0,str(pathlib.Path(__file__).resolve().parents[1]/"experiments"))
from todo_promotion import preview
class PromotionTest(unittest.TestCase):
    def test_selection_review_and_duplicate_mapping(self):
        tasks=[{"id":"a","title":"private text"},{"id":"b","title":"reviewed","disclosure_review":"approved"},{"id":"c","title":"not selected"}]
        self.assertEqual(preview(tasks,["a","b"],{}),[{"id":"a","state":"blocked_review"},{"id":"b","state":"preview_only","title":"reviewed"}])
        self.assertEqual(preview(tasks,["b"],{"b":"issue-1"}),[{"id":"b","state":"already_published"}])
        with self.assertRaises(ValueError):preview(tasks,["b","b"],{})
        with self.assertRaises(ValueError):preview(tasks,["missing"],{})
if __name__=="__main__":unittest.main()
