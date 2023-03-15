State machine replication (SMR) allows nodes to
jointly maintain a consistent ledger, even when a part of nodes
are Byzantine. To defend against and/or limit the impact of
attacks launched by Byzantine nodes, there have been blocks
that combine reputation mechanisms to SMR, where each node
has a reputation value based on its historical behaviours, and the
node’s voting power will be proportional to its reputation. Despite
the promising features of reputation-based SMR, existing studies
do not provide formal treatment on the reputation mechanism on
SMR protocols, including the types of behaviours affecting the
reputation, the security properties of the reputation mechanism,
and the extra security properties of SMR using reputation
mechanisms.
In this paper, we provide the first formal study on the
reputation-based SMR. We define the security properties of the
reputation mechanism w.r.t. these misbehaviours. Based on the
formalisation of the reputation mechanism, we formally define the
reputation-based SMR, and identify a new property reputation-
consistency that is necessary for ensuring reputation-based SMR’s
safety. We then design a simple reputation mechanism that
achieves all security properties in our formal model. To demon-
strate the practicality, we combine our reputation mechanism to
the Sync-HotStuff SMR protocol, yielding a simple and efficient
reputation-based SMR at the cost of only an extra ∆ in latency,
where ∆ is the maximum delay in synchronous networks.
****************************************************************
This paper have been accepted by 2022 IEEE 21st International Symposium on Network Computing and Applications (NCA).
****************************************************************
QUICK START
1. sh quicktest.sh.
2. Replace n.startConsensusTimer() in ../consensus/command.go with different misbehaviors in attackInjection.go to simulate different misbehaviors.
3. ../tools adjust block size, node num.

