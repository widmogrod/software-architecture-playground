package projection

import "testing"

func TestProjection(t *testing.T) {
	_ = "SELECT 'Hello World' AS Greeting"
	_ = `SELECT children[0].fname AS child_name
    FROM tutorial
       WHERE fname='Dave'`

	_ = `SELECT fname, email
    FROM tutorial 
        WHERE email LIKE '%@yahoo.com'`
	_ = `SELECT fname, children
    FROM tutorial 
        WHERE ANY child IN tutorial.children SATISFIES child.age > 10  END`
	_ = `SELECT fname, email, children
    FROM tutorial 
        WHERE ARRAY_LENGTH(children) > 0 AND email LIKE '%@gmail.com'`

	// Specific primary keys within a bucket can be queried using the USE KEYS clause.
	_ = `SELECT fname, email
    FROM tutorial 
        USE KEYS ["dave", "ian"]`

	// Similar to filtering documents with the WHERE clause, we can filter groups with the HAVING clause.
	//Here we filter to only include groups with more than 1 member.
	_ = `SELECT relation, COUNT(*) AS count
    FROM tutorial
        GROUP BY relation
            HAVING COUNT(*) > 1`

	_ = "SELECT sessionID, count() as winds, count() as draws, winds + draws as total FROM sessions GROUP BY sessionID"

}
