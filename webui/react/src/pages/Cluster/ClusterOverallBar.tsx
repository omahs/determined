import Message from 'determined-ui/Message';
import Spinner from 'determined-ui/Spinner';
import { ShirtSize } from 'determined-ui/Theme';
import { Loadable } from 'determined-ui/utils/loadable';
import React, { useMemo } from 'react';

import Section from 'components/Section';
import SlotAllocationBar from 'components/SlotAllocationBar';
import clusterStore from 'stores/cluster';
import { ResourceType } from 'types';
import { getSlotContainerStates } from 'utils/cluster';
import { useObservable } from 'utils/observable';

export const ClusterOverallBar: React.FC = () => {
  const overview = useObservable(clusterStore.clusterOverview);
  const agents = useObservable(clusterStore.agents);

  const cudaSlotStates = useMemo(() => {
    return getSlotContainerStates(
      (Loadable.isLoaded(agents) && agents.data) || [],
      ResourceType.CUDA,
    ); // can't use a const here due to types
  }, [agents]);

  const rocmSlotStates = useMemo(() => {
    return getSlotContainerStates(
      (Loadable.isLoaded(agents) && agents.data) || [],
      ResourceType.ROCM,
    );
  }, [agents]);

  const cpuSlotStates = useMemo(() => {
    return getSlotContainerStates(
      (Loadable.isLoaded(agents) && agents.data) || [],
      ResourceType.CPU,
    );
  }, [agents]);

  return (
    <Section hideTitle title="Overall Allocation">
      <Spinner
        conditionalRender
        spinning={Loadable.isNotLoaded(agents) || Loadable.isNotLoaded(overview)}>
        {Loadable.isLoaded(agents) && Loadable.isLoaded(overview) ? ( // This is ok as the Spinner has conditionalRender active
          <>
            {overview.data.CUDA.total + overview.data.ROCM.total + overview.data.CPU.total === 0 ? (
              <Message icon="warning" title="No connected agents." />
            ) : null}
            {overview.data.CUDA.total > 0 && (
              <SlotAllocationBar
                resourceStates={cudaSlotStates}
                showLegends
                size={ShirtSize.Large}
                title={`Compute (${ResourceType.CUDA})`}
                totalSlots={overview.data.CUDA.total}
              />
            )}
            {overview.data.ROCM.total > 0 && (
              <SlotAllocationBar
                resourceStates={rocmSlotStates}
                showLegends
                size={ShirtSize.Large}
                title={`Compute (${ResourceType.ROCM})`}
                totalSlots={overview.data.ROCM.total}
              />
            )}
            {overview.data.CPU.total > 0 && (
              <SlotAllocationBar
                resourceStates={cpuSlotStates}
                showLegends
                size={ShirtSize.Large}
                title={`Compute (${ResourceType.CPU})`}
                totalSlots={overview.data.CPU.total}
              />
            )}
          </>
        ) : undefined}
      </Spinner>
    </Section>
  );
};
