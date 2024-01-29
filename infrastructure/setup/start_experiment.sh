# todo idea for nodeid stuff: when starting, do "export \$NODEID=0; ./main"
# at least this works:
# `gcloud compute ssh producer-instance-0 --zone=europe-west10-a --command='export A=1; env'`
# so the programs should be able to also see that environment variable??

# todo quorum queue configure replica count
# todo rabbitmq configure memory stuff 40% -> 70%
