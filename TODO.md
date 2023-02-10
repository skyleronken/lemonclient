- Server
    - /lg/test
    - /lg
- Adapter
    - /lg/adapter/{adapter} GET/POST
    - /lg/adapter/{adapter}/{job_uuid} POST
- Task
    - /lg/task/{job_uuid}/{task_uuid} GET
    - /lg/task/{job_uuid}/{task_uuid} HEAD
    - /lg/task/{job_uuid}/{task_uuid} DELETE
    - /lg/task/{job_uuid}/{task_uuid} POST
- Job
    - /graph GET [bulk]

        [
        {
            "graph": "af09a42b-a97b-11ed-925d-0242ac110002",
            "id": "af09a42b-a97b-11ed-925d-0242ac110002",
            "meta": {
                "roles": {
                    "bob": {
                        "reader": true,
                        "writer": false
                    }
                },
                "priority": 100,
                "enabled": true
            },
            "size": 57344,
            "nodes_count": 2,
            "edges_count": 1,
            "maxID": 15,
            "created": "2023-02-10T19:47:00.106961Z"
        },
        ]

    - /graph/{uuid} HEAD
    - /graph/{uuid} GET
    - /graph/{uuid} POST
    - /graph/{uuid} PUT
    - /graph/{uuid} DELETE
    - /graph/{uuid}/meta GET
    - /graph/{uuid}/meta PUT
    - /graph/{uuid}/seeds GET
    - /graph/{uuid}/status GET
    - /graph/{uuid}/status HEAD
    - /reset/{uuid} PUT
    - /lg/config/{job_uuid} GET
    - /lg/config/{job_uuid} POST
    - /lg/config/{job_uuid}/adapter GET
    - /lg/config/{job_uuid}/adapter POST
    - /lg/task/{job_uuid} GET
    - /lg/task/{job_uuid} POST
    - /lg/delta/{job_uuid} GET/POST


- Add graph-specific endpoint parameters for user and role
- Add bulk endpoint query parameters
- d3 stuff
- key value stuff
- /lg/test
- add 'id' to edge and node
    - /graph/{uuid}/edge/{ID}
    - /graph/{uuid}/node/{ID}
- Unmarshall Job
