import sqlite3
import sys

import pandas as pd
import numpy as np
import seaborn as sns
import matplotlib.pyplot as plt

from sqlite3 import Error


#
# parameters shared between plots
#

style   = "ticks"
context = "notebook"
colors7  = [
    "#fde725",
    "#90d743",
    "#35b779",
    "#21918c",
    "#31688e",
    "#443983",
    "#440154",
]
colors3 = [
    "#fde725",
    "#35b779",
    "#440154",
]

sns.set(
    style=style,
    context=context
)

out_dir = "./plots/"


#
# database access
# 

db_file    = "./csb.db"
connection = sqlite3.connect(db_file)
cursor     = connection.cursor()


#
# throughput
#

# aggregate the data
# background: the `request` table in the DB has one entry for each successfully delivered message
# this aggregation creates a new table (if it doesn't exist already) with the amount of successfully delivered messages per second

query = """
CREATE TABLE IF NOT EXISTS throughput AS
SELECT count(tProducer) as nMessages, tConsumer, run, nBrokers
FROM (
    SELECT tProducer / 1000 as tProducer, tConsumer / 1000 as tConsumer, run, nBrokers
    FROM request
)
GROUP BY tConsumer, run, nBrokers
ORDER BY tConsumer ASC;
"""

try:
    cursor.execute(query)
    print("table 'throughput' created successfully")

except Error as e:
    print(e)
    sys.exit(1)
    

# plot: 
# throughput across broker configurations and runs
 
throughput = pd.read_sql_query("SELECT * FROM throughput", connection)

# data filtering
# 1) there might and small amount of rows where the monitor was interrupted while writing the line
throughput = throughput[throughput['tConsumer'] >= 1e9]
# 2) convert unix timestamps to relative timestamps since start of experiment run
throughput['experiment_timestamp'] = throughput.groupby(['run', 'nBrokers'])['tConsumer'].transform(lambda x: x - x.min())
# 3) filter out the data after the experiment was over
throughput = throughput[throughput['experiment_timestamp'] < 3_660_000]

# data transformation
# apply smoothing
window_size = 5
throughput['nMessages_smoothed'] = throughput.groupby(['run', 'nBrokers'])['nMessages'].transform(lambda x: x.rolling(window=window_size).mean())

# create a grid with a subplot for each broker configuration and experiment run
#g_throughput = sns.FacetGrid(throughput, col="run", row="nBrokers", hue="nBrokers", margin_titles=True, height=3, aspect=1.5)
g_throughput = sns.FacetGrid(throughput, col="run", margin_titles=True, height=5, aspect=0.8)
g_throughput.map_dataframe(sns.lineplot, x="experiment_timestamp", y="nMessages_smoothed", hue="nBrokers", errorbar="ci", markers=True, dashes=False, palette=colors3[::-1])
g_throughput.set_axis_labels("Time [m]", "Throughput [msg/s]")
g_throughput.set_titles(col_template="Run {col_name}")
g_throughput.set(xticks=[0, 900, 1800, 2700, 3600], xticklabels=['0', '15', '30', '45', '60'])
g_throughput.set(ylim=(0, 14000))
g_throughput.add_legend(title="Nodes")

plt.savefig(out_dir + "throughput.pdf")


# plot:
# throughput ECDF
# leave out the legend here because it will be on the same slide as the next plot,
# and they will share one legend
# => maybe add back it if the plot makes it into the report

# get the appropriate data without the `run` column
df = pd.read_sql_query("""
    SELECT nMessages as amount, tConsumer as timestamp, nBrokers as nodes
    FROM throughput;
""", connection)

# get max throughput to find correct xlim for ECDF plot
maxThroughput = pd.read_sql_query("""
    SELECT max(nMessages) as max
    FROM throughput
    LIMIT 1;
""", connection).iloc[0, 0]

# use 1x1 grid to make the legend similar to the other plots
g_ecdf = sns.FacetGrid(df, hue="nodes", height=5, aspect=1, palette=colors3[::-1])
g_ecdf.map_dataframe(sns.ecdfplot, "amount", linewidth=2)
g_ecdf.set_axis_labels("Throughput [msg/s]", "ECDF")
g_ecdf.set(xlim=(0, maxThroughput))
g_ecdf.set(ylim=(0,1))
g_ecdf.add_legend(title="Nodes")

g_ecdf.savefig(out_dir + "throughput_ecdf.pdf")


# plot:
# throughput KDE
# => use throughput dataframe from before

g_kde = sns.FacetGrid(throughput, hue="nBrokers", height=5, aspect=(4/3), palette=colors3[::-1]) 
g_kde.map(sns.kdeplot, "nMessages", fill=True, linewidth=2)
g_kde.set_axis_labels("Throughput [msg/s]", "Density")
g_kde.set(xlim=(0, maxThroughput))
g_kde.set(ylim=(0, 0.0007))
g_kde.add_legend(title="Nodes")

g_kde.savefig(out_dir + "throughput_kde.pdf")


#
# CPU utilization
#

for runId in range(1, 4):

    cpu = pd.read_sql_query(f"""
        SELECT run, brokerId, nBrokers as nodes, timestamp, idlep
        FROM cpu
        WHERE run = {runId};
    """, connection)

    # filter data
    cpu = cpu[cpu['timestamp'] >= 1e9]
    cpu['experiment_timestamp'] = cpu.groupby(['run', 'nodes'])['timestamp'].transform(lambda x: x - x.min())
    cpu['experiment_timestamp_seconds'] = cpu['experiment_timestamp'] / 1000 

    cpu = cpu[cpu['experiment_timestamp'] < 3_660_000]

    # transform data
    # idle % -> busy %
    cpu['utilization'] = 100 - cpu['idlep']

    # smooth data
    rolling_window = 15
    cpu['utilization'] = cpu.groupby(['run', 'nodes', 'brokerId'])['utilization'].transform(lambda x: x.rolling(window=rolling_window).mean())

    # create the plot
    g_cpu = sns.FacetGrid(cpu, col="nodes", margin_titles=True, height=5, aspect=0.8)
    g_cpu.map_dataframe(sns.lineplot, x="experiment_timestamp_seconds", y="utilization", hue="brokerId", markers=True, dashes=False, palette=colors7[::-1])
    g_cpu.set_axis_labels("Time [m]", "CPU Utilization [%]")
    g_cpu.set_titles(col_template="{col_name} Nodes")
    g_cpu.set(xticks=[0, 900, 1800, 2700, 3600], xticklabels=['0', '15', '30', '45', '60'])
    g_cpu.set(ylim=(0, 100))
    g_cpu.add_legend(title="Broker ID")

    plt.savefig(out_dir + f"cpu-run-{runId}.pdf")


#
# memory usage
#

for runId in range(1, 4):

    memory = pd.read_sql_query(f"""
        SELECT run, brokerId, nBrokers as nodes, timestamp, freep
        FROM memory
        WHERE run = {runId};
    """, connection)

    # filter data
    memory = memory[memory['timestamp'] >= 1e9]
    memory['experiment_timestamp'] = memory.groupby(['run', 'nodes'])['timestamp'].transform(lambda x: x - x.min())
    memory['experiment_timestamp_seconds'] = memory['experiment_timestamp'] / 1000 

    memory = memory[memory['experiment_timestamp'] < 3_660_000]

    # transform data
    memory['usage'] = 100 - memory['freep']

    # smooth data
    rolling_window = 1
    memory['usage'] = memory.groupby(['run', 'nodes', 'brokerId'])['usage'].transform(lambda x: x.rolling(window=rolling_window).mean())

    # create the plot
    g_memory = sns.FacetGrid(memory, col="nodes", margin_titles=True, height=5, aspect=0.8)
    g_memory.map_dataframe(sns.lineplot, x="experiment_timestamp_seconds", y="usage", hue="brokerId", markers=True, dashes=False, palette=colors7[::-1])
    g_memory.set_axis_labels("Time [m]", "Memory Usage [%]")
    g_memory.set_titles(col_template="{col_name} Nodes")
    g_memory.set(xticks=[0, 900, 1800, 2700, 3600], xticklabels=['0', '15', '30', '45', '60'])
    g_memory.set(ylim=(0, 100))
    g_memory.add_legend(title="Broker ID")

    plt.savefig(out_dir + f"memory-run-{runId}.pdf")


#
# network usage: bytes transmitted
#

for runId in range(1, 4):

    txbytes = pd.read_sql_query(f"""
        SELECT run, brokerId, nBrokers as nodes, timestamp, txbytes
        FROM network
        WHERE run = {runId};
    """, connection)

    # filter data
    txbytes = txbytes[txbytes['timestamp'] >= 1e9]
    txbytes['experiment_timestamp'] = txbytes.groupby(['run', 'nodes'])['timestamp'].transform(lambda x: x - x.min())
    txbytes['experiment_timestamp_seconds'] = txbytes['experiment_timestamp'] / 1000 

    txbytes = txbytes[txbytes['experiment_timestamp'] < 3_660_000]

    # transform data
    # bytes -> megabits
    txbytes["TxBytes"] = txbytes["TxBytes"] * 8 / 1e6

    # smooth data
    rolling_window = 10
    txbytes['TxBytes'] = txbytes.groupby(['run', 'nodes', 'brokerId'])['TxBytes'].transform(lambda x: x.rolling(window=rolling_window).mean())

    # create the plot
    g_txbytes = sns.FacetGrid(txbytes, col="nodes", margin_titles=True, height=5, aspect=0.8)
    g_txbytes.map_dataframe(sns.lineplot, x="experiment_timestamp_seconds", y="TxBytes", hue="brokerId", markers=True, dashes=False, palette=colors7[::-1])
    g_txbytes.set_axis_labels("Time [m]", "Network Usage [Mbit/s Transmitted]")
    g_txbytes.set_titles(col_template="{col_name} Nodes")
    g_txbytes.set(xticks=[0, 900, 1800, 2700, 3600], xticklabels=['0', '15', '30', '45', '60'])
    g_txbytes.set(ylim=(0, 600))
    g_txbytes.add_legend(title="Broker ID")

    plt.savefig(out_dir + f"txbytes-run-{runId}.pdf")


#
# network usage: bytes received
#


for runId in range(1, 4):

    rxbytes = pd.read_sql_query(f"""
        SELECT run, brokerId, nBrokers as nodes, timestamp, RxBytes
        FROM network
        WHERE run = {runId};
    """, connection)

    # filter data
    rxbytes = rxbytes[rxbytes['timestamp'] >= 1e9]
    rxbytes['experiment_timestamp'] = rxbytes.groupby(['run', 'nodes'])['timestamp'].transform(lambda x: x - x.min())
    rxbytes['experiment_timestamp_seconds'] = rxbytes['experiment_timestamp'] / 1000 

    rxbytes = rxbytes[rxbytes['experiment_timestamp'] < 3_660_000]

    # transform data
    # bytes -> megabits
    rxbytes["RxBytes"] = rxbytes["RxBytes"] * 8 / 1e6

    # smooth data
    rolling_window = 10
    rxbytes['RxBytes'] = rxbytes.groupby(['run', 'nodes', 'brokerId'])['RxBytes'].transform(lambda x: x.rolling(window=rolling_window).mean())

    # create the plot
    g_rxbytes = sns.FacetGrid(rxbytes, col="nodes", margin_titles=True, height=5, aspect=0.8)
    g_rxbytes.map_dataframe(sns.lineplot, x="experiment_timestamp_seconds", y="RxBytes", hue="brokerId", markers=True, dashes=False, palette=colors7[::-1])
    g_rxbytes.set_axis_labels("Time [m]", "Network Usage [Mbit/s Received]")
    g_rxbytes.set_titles(col_template="{col_name} Nodes")
    g_rxbytes.set(xticks=[0, 900, 1800, 2700, 3600], xticklabels=['0', '15', '30', '45', '60'])
    g_rxbytes.set(ylim=(0, 150))
    g_rxbytes.add_legend(title="Broker ID")

    plt.savefig(out_dir + f"rxbytes-run-{runId}.pdf")


#
# the end
# 

#plt.show()
connection.close()
