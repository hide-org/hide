package jsonc

import "testing"

func TestToJSON(t *testing.T) {
	json := `
  {  //	hello
    "c": 3,"b":3, // jello
    /* SOME
       LIKE
       IT
       HAUT */
    "d\\\"\"e": [ 1, /* 2 */ 3, 4, ],
  }`
	expect := `
  {    	     
    "c": 3,"b":3,         
           
           
         
              
    "d\\\"\"e": [ 1,         3, 4  ] 
  }`
	out := string(ToJSON([]byte(json)))
	if out != expect {
		t.Fatalf("expected '%s', got '%s'", expect, out)
	}
}

func TestToJSON2(t *testing.T) {
	json := "{\n  \"name\": \"Sample Dev Container\",\n  \"image\": \"pttest:local\",\n  \"settings\": {},\n  \"extensions\": [],\n  \"postCreateCommand\": \"echo Welcome to the dev container\",\n  \"forwardPorts\": [3000],\n  \"remoteUser\": \"vscode\"\n}"
	expect := json

	out := string(ToJSON([]byte(json)))
	if out != expect {
		t.Fatalf("expected '%s', got '%s'", expect, out)
	}
}
