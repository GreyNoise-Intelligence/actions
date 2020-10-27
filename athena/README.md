# Athena Github Action

This Github action scans a directory for Athena SQL files and executes them. These files should have statements to create views or other schema objects. 

## Inputs

#### `path`

**Required** The path to scan for SQL files that contain Athena views. Only files that end with a `.sql` extension will be processed. 

#### `database`

**Required** The Athena database to create views in

#### `region`

**Optional** The AWS session region - defaults to `us-east-1`.

#### `workgroup`

**Optional** The AWS Athena workgroup to output views to.

## Outputs

#### `result`

The CLI output from executing Athena queries.

## Example usage

```yaml
on: [push]

jobs:
  refresh_views:
    runs-on: ubuntu-latest
    name: Refresh Athena views
    steps:
    - name: Generate Athena Views
      id: athena_views
      uses: nayyara-samuel/actions/athena@master
      with:
        path: infrastructure/bi/schema
        database: bi_db
      env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}    
    - name: Example
      run: echo "The result was ${{ steps.athena_views.outputs.result }}"
```  

## Testing Locally

Requires AWS credentials set and Athena database to exist. 

```bash 
export INPUT_PATH=xxx
export INPUT_DATABASE=xxx
make run 
```

##### Example Output

<img src="local.png" alt="local-run"/>
