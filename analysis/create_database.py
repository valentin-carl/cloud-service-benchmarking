import csv
import os
import re
import sqlite3
from sqlite3 import Error

import pandas as pd


# TODO adjust this if running on a different computer
data    = "/Users/valentincarl/Nextcloud/tubcloud/csb/raw-data/"
db_file = "./csb.db"

print(os.listdir(data))

# create database
connection = sqlite3.connect(db_file)
assert connection is not None
if connection:
    print(sqlite3.version)

# create tables
def create_table(table_name: str, columns: dict) -> None:
    query = "CREATE TABLE {} (\n".format(table_name)
    for column_name in columns.keys():
        query += "\t{} {},\n".format(column_name, columns[column_name])
    query = query.rstrip(",\n") + "\n);"
    return query

# yes, it would've been easier to just write the queries
tables = [
    {
        "table_name": "request",
        "columns": {
            "tProducer": "INTEGER",
                "tConsumer": "INTEGER",
                "run": "INTEGER",
                "nBrokers": "INTEGER",
                "consumerId": "INTEGER",
        }
    },
    {
        "table_name": "cpu",
        "columns": {
            "run": "INTEGER",
            "brokerId": "INTEGER",
            "timestamp": "INTEGER",
            "userp": "REAL",
            "systemp": "REAL",
            "idlep": "REAL"
        }
    },
    {
        "table_name": "memory",
        "columns": {
            "run": "INTEGER",
            "brokerId": "INTEGER",
            "timestamp": "INTEGER",
            "freep": "REAL"
        }
    },
    {
        "table_name": "network",
        "columns": {
            "run": "INTEGER",
            "brokerId": "INTEGER",
            "timestamp": "INTEGER",
            "RxBytes": "INTEGER",
            "TxBytes": "INTEGER",
        }
    }
]

c = connection.cursor()

for table in tables:
    query = create_table(table["table_name"], table["columns"])
    #print(query)
    try:
        c.execute(query)
    except Error as e:
        print(e)

# verify that the tables were created successfully
print(c.execute("SELECT name FROM sqlite_master WHERE type='table';").fetchall())

# load the request data into the database
# executing this part took ~5 minutes on my computer, 
# the resulting DB file is 6.26 GB large, so this probably
# depends a lot on the hard drive's write speed

chunk_size = 10_000

def extract(dir_name):
    pattern = r"broker-(\d+)-run-(\d+)-(cpu|memory|network)\.csv"
    match = re.match(pattern, dir_name)
    if match:
        brokerId = int(match.group(1))
        run = int(match.group(2))
        return brokerId, run # FIXME run is wrong!! use get_run_nbrokers() instead
    else:
        raise Error(f"error while trying to find 'brokerId' and 'run' in '{dir_name}'")
        
def get_run_nbrokers(item):
    pattern = r"run-(\d+)-nbrokers-(\d)+"
    match = re.match(pattern, item)
    assert match is not None
    return match.group(1), match.group(2)

for item in os.listdir(data):

    # skip non-directories
    path = os.path.join(data, item)
    if os.path.isfile(path):
        continue
    print(f"- {item}")

    # go through sub-directories and find measurements 
    for subdir in os.listdir(path):

        sub = os.path.join(path, subdir)
        printt = lambda s : print(f"\t- {s}")

        match subdir:

            case "broker":
                
                # print directory contents
                measurement_files = [s for s in os.listdir(os.path.join(sub, "data")) if s[0] != '.']
                printt(f"broker/data:\t{measurement_files}")

                for filename in measurement_files:

                    # depending on the measurement type, different columns have to be extracted from the file
                    filepath = os.path.join(sub, "data", filename)
                    pattern = r"broker-\d+-run-\d+-(cpu|memory|network)\.csv"
                    match = re.match(pattern, filename)
                    measurement_type = match.group(1)
                    match measurement_type:
                        case "cpu":
                            cols = ["timestamp", "userp", "systemp", "idlep"]
                        case "memory":
                            cols = ["timestamp", "freep"]
                        case "network":
                            cols = ["timestamp", "RxBytes", "TxBytes"]
                        case _:
                            print("no match!")

                    # insert the data into the database
                    df_chunks = pd.read_csv(
                        filepath, 
                        usecols=cols,
                        chunksize=chunk_size
                    )
                    for chunk in df_chunks:
                        # add additional columns
                        brokerId, run = extract(filename)
                        chunk["brokerId"] = brokerId
                        #chunk["run"] = run # stuoopid mistake: all files are named ...run-0... -.- => use runId from directory instead
                        chunk["run"], _ = get_run_nbrokers(item)
                        # using "with" turns this into a transaction:
                        # https://blog.rtwilson.com/a-python-sqlite3-context-manager-gotcha/
                        with connection:
                            chunk.to_sql(measurement_type, connection, index=False, if_exists="append")

            case "consumer":

                # print directory contents
                measurement_files = [s for s in os.listdir(os.path.join(sub, "out")) if s[0] != '.']
                printt(f"consumer/out:\t{measurement_files}")

                for filename in measurement_files:

                    filepath = os.path.join(sub, "out", filename)

                    # get run id
                    run, nbrokers = get_run_nbrokers(item)

                    # get consumer node id
                    pattern = r"experiment-run-\d+-node-(\d+).csv"
                    match = re.match(pattern, filename)
                    assert match is not None
                    consumerId = int(match.group(1))

                    # load + insert into DB
                    cols = ["tProducer", "tConsumer"]
                    df_chunks = pd.read_csv(
                        filepath,
                        usecols=cols,
                        chunksize=chunk_size
                    )
                    for chunk in df_chunks:
                        # add additional columns
                        chunk["run"] = run
                        chunk["nBrokers"] = nbrokers
                        chunk["consumerId"] = consumerId
                        # insertion time!
                        with connection:
                            chunk.to_sql("request", connection, index=False, if_exists="append")

            case _:
                if subdir[0] != '.':
                    printt(subdir)

    print()

connection.close()
