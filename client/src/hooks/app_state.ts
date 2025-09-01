import { useState } from 'react';

export type ModalType = 'admin' | null;

function UseAppState() {
    const [activeModal, setActiveModal] = useState<ModalType>(null);

    return { setActiveModal, activeModal };
}

export default UseAppState;
