We answer the following questions during this evaluation:
1.How does Reputation-based SMR perform in best-case(latency and throughout)?
2.How does Reputation-based SMR perform in worst-case((latency and throughout))?
3. Can Reputation-based SMR ensure the accuracy of node reputation under accountable misbehaviours?
4. Can Reputation-based SMR ensure the reputation consistency?
5. Can Reputation-based SMR guarantee flash-attack resistance like Repucoin and Guru?
6. How does Reputation-based SMR compare to the baselines (Sync-HotStuff,RepuCoin) in the performance(latency and throughout)?
Evaluation method(approximate ):
First, we measure the performance in best-case and worst-case to prove that reputation-based SMR has the same performance in the different cases(i.e., best-case and worst-case).
Then, we prove the accuracy of reputation-based SMR by verifying the correctness of each node's reputation score changes in different cases and we verify the reputation consistency by the same transcript of each honest node under the same epoch.
In addition, we instantiate each parameter(e.g., ) in the reputation function, and use the experimental data to illustrate the specific cost of the number of nodes and time that the adversary pays to successfully initiate flash-attack.
Finally, under the same setting, we compare the performance of Reputation-based SMR with Sync-HotStuff to prove that Reputation-based SMR achieves slashing support and flash-attack resistance at the cost of only one , at the same time, we also compared the performance of Reputation-based SMR with RepuCoin.
