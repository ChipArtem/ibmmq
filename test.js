import ibmq from "k6/x/ibmq";

const conn = ibmq.connect();

export const setup = () => {
    console.log("setup!");
};

export default () => {
    ibmq.write();
    ibmq.read();
};

export const teardown = () => {
    console.log("teardown");
};