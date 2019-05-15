from mesa import Model
from mesa.time import SimultaneousActivation
import agent
import numpy as np
from mesa.datacollection import DataCollector
import logging as log

class Beacon_Model(Model):
    """The model"""
    def __init__(self, nodes, ticket_distribution, active_group_threshold, 
    group_size, max_malicious_threshold, group_expiry, 
     node_failure_percent, node_death_percent,
    signature_delay, min_nodes, node_connection_delay, node_mainloop_connection_delay, log_filename, run_number):
        self.num_nodes = nodes
        self.schedule = SimultaneousActivation(self)
        self.relay_request = False
        self.active_groups = []
        self.active_nodes = []
        self.active_group_threshold = active_group_threshold # number of groups that will always be maintained in an active state
        self.max_malicious_threshold = max_malicious_threshold # threshold above which a signature is deemed to be compromised, typically 51%
        self.group_size = group_size
        self.ticket_distribution = ticket_distribution
        self.newest_id = 0
        self.newest_group_id = 0
        self.newest_signature_id = 0
        self.group_expiry = group_expiry
        self.bootstrap_complete = False # indicates when the initial active group list bootstrap is complete
        self.group_formation_threshold = min_nodes # min nodes required to form a group
        self.timer = 0
        self.unsuccessful_signature_events = []
        self.signature_delay = signature_delay
        self.datacollector = DataCollector(
            agent_reporters={"Ownership_distribution": "ownership_distr"})  # Collect ownership distributions for groups

        #create log file
        log.basicConfig(filename=log_filename + str(run_number), filemode='w', format='%(name)s - %(levelname)s - %(message)s')


        #create nodes
        for i in range(nodes):
            node = agent.Node(i, i, self, self.ticket_distribution[i], node_failure_percent, node_death_percent, node_connection_delay, node_mainloop_connection_delay)
            self.newest_id = i
            self.schedule.add(node)
        self.newest_id +=1


    def step(self):
        '''Advance the model by one step'''
        log.debug("step # = " + str(self.timer))
 
        #check how many active nodes are available
        self.refresh_connected_nodes_list()
        log.debug("Number of nodes in the forked state = " + str(len(self.active_nodes)))

        #bootstrap active groups as nodes become available. Can only happen once enough nodes are online
        temp_bootstrap_groups = []
        if self.bootstrap_complete == False:
            log.debug("bootstrapping active groups")
            if len(self.active_nodes)>=self.group_formation_threshold:
                for i in range(self.active_group_threshold):
                    temp_bootstrap_groups.append(self.group_registration())
                self.bootstrap_complete = True
            self.active_groups = temp_bootstrap_groups

        #check how many active groups are available
        self.refresh_active_group_list()
        
        #generate relay requests
        self.relay_request = np.random.choice([True,False]) # make this variable so it can be what-if'd
        log.debug("relay request recieved? = "+ str(self.relay_request))

        if self.relay_request:
            try:
                log.debug('     selecting group at random')
                try:
                    # pick an active group from the active group list and create a signature object
                    signature = agent.Signature(self.newest_id, self.newest_signature_id, self, self.active_groups[np.random.randint(len(self.active_groups))]) 
                except Exception as e: print(e)
                self.schedule.add(signature)
            except:
                log.debug('     no active groups available')

            log.debug('     registering new group')
            self.group_registration()
        else:
            log.debug("     No relay request")
        self.timer += 1

        #advance the agents
        self.schedule.step()

    def group_registration(self):
        ticket_list = []
        group_members = []

        if len(self.active_nodes)<self.group_formation_threshold: 
            log.debug("             Not enough nodes to register a group")

        else:
            # make each node generate tickets and save them to a list
            max_tickets = int(max(self.ticket_distribution))
            for node in self.active_nodes:
                adjusted_ticket_list = []
                node.generate_tickets()
                adjusted_ticket_list = np.concatenate([node.ticket_list,np.ones(int(max_tickets)-len(node.ticket_list))])  #adds 2's the ends of the list so that the 2D array can have equal length rows
                ticket_list.append(adjusted_ticket_list)

            #iteratively add group members by lowest value
            while len(group_members) <= self.group_size:

                min_index = np.where(ticket_list == np.min(ticket_list)) # find the index of the minimum value in the array
                for i,index in enumerate(min_index[0]): #if there are repeated values, iterate through and add the indexes to the group
                    group_members.append(self.active_nodes[index])
                    ticket_list[index][min_index[1][i]] = 2 # Set the value of the ticket to a high value so it doesn't get counted again
            
            #create a group agent which can track expiry, sign, etc
            group_object = agent.Group(self.newest_id, self.newest_group_id, self, group_members, self.group_expiry)


            #add group to schedule
            self.schedule.add(group_object)

            #add group to active group list
            self.active_groups.append(group_object)
            
            return group_object

    def refresh_active_group_list(self):
        temp_list = []

        for group in self.active_groups:
            if group.status == "Active":
                temp_list.append(group)
        
        self.active_groups = temp_list

    def refresh_connected_nodes_list(self):
        log.debug("refreshing active nodes list")
        temp_active_node_list = []
        for agent in self.schedule.agents:
            if agent.type == "node":
                if agent.mainloop_status == "forked": 
                    temp_active_node_list.append(agent) #adds the node to the active list only if it is in the forked state
        self.active_nodes = temp_active_node_list



    










        

        







